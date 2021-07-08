#!/usr/bin/env python

import os
import sys
import yaml

yamltypes2go = {
    'integer': 'int',
    'int64': 'int64',
    'boolean': 'bool',
    'string': 'string'
}

# type name mapping for structs and sub structs found within a CRD.
# we want some of the sub structs to be visible to calling code,
# so that we can return instances/arrays of these types.
# _spec and _status are top level tokens, selected to avoid collision
# with naturally occurring tokens like spec,
# because for example the MSV status has a Spec field.
typeNameMaps = {
    'mayastorvolume': {
        '_spec': 'MayastorVolumeSpec',
        '_status': 'MayastorVolumeStatus',
        'replicas': 'Replica',
        'children': 'NexusChild',
        'nexus': 'Nexus',
        'targetnodes': 'TargetNodes',
    },
    'mayastorpool': {
        '_spec': 'MayastorPoolSpec',
        '_status': 'MayastorPoolStatus',
    },
    'mayastornode': {
        '_spec': 'MayastorNodeSpec',
        'grpcendpoint': 'GrpcEndpoint',
    },
}

# mapping for CRD type name with custom capitalization
crdTypeMap = {
    'mayastorvolume': 'MayastorVolume',
    'mayastorpool':  'MayastorPool',
    'mayastornode':  'MayastorNode',
}

template = """package v1alpha1

import metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

{defs}

type {crdType} struct {{
        metaV1.TypeMeta   `json:",inline"`
        metaV1.ObjectMeta `json:"metadata,omitempty"`

        Spec   {specType}   `json:"spec"`
        Status {statusType} `json:"status"`
}}

type {crdListType} struct {{
        metaV1.TypeMeta `json:",inline"`
        metaV1.ListMeta `json:"metadata,omitempty"`

        Items []{crdType} `json:"items"`
}}
"""


def parseYamlArray(ydict, goStructs, name):
    """
    parse the yaml dictiontary (ydict) array entry and
    populate the list of go structs (goStructs)
    """
    typ = ydict['type']
    format = ydict.get('format')
    if typ in ['array']:
        return []
    elif typ in ['object']:
        return [parseYaml(ydict, goStructs, name)]
    else:
        if format is None:
            goTyp = yamltypes2go[typ]
        else:
            goTyp = yamltypes2go[format]
        return [goTyp]


def parseYaml(ydict, goStructs, name):
    """
    parse the yaml dictiontary (ydict) and
    populate the list of go structs (goStructs)
    """
    if ydict['type'] not in ['object', 'array']:
        return ydict['type']
    elif ydict['type'] == 'object':
        obj = {}
        for k, v in ydict['properties'].items():
            typ = v['type']
            format = v.get('format')
            if typ == 'object':
                obj[k] = parseYaml(v, goStructs, k)
            elif typ == 'array':
                obj[k] = parseYamlArray(v['items'], goStructs, k)
            else:
                if format is None:
                    goTyp = yamltypes2go[typ]
                else:
                    goTyp = yamltypes2go[format]
                obj[k] = goTyp
        goStructs.append({name: obj})
        return name
    else:
        raise Exception('unsupported type {}'.format(ydict['type']))


def capitalize(s):
    """
    Capitalise just the 1st element without lower casing the other elements
    """
    return s[0:1].capitalize() + s[1:]


class GenGoCrd():
    def __init__(self, yamlfile):
        self.key = os.path.basename(yamlfile).split('.')[0]
        self.typeNameMap = typeNameMaps.get(self.key)
        if self.typeNameMap is None:
            raise Exception('unknown yamlfile {}'.format(self.key))

        with open(yamlfile, mode='r') as fp:
            contents = fp.read()
            ymls = list(yaml.safe_load_all(contents))
            if len(ymls) != 1:
                raise Exception('{} contains multiple streams'.format(
                                yamlfile))

            schema = ymls[0]['spec']['versions'][0]['schema']
            spec = schema['openAPIV3Schema']['properties']['spec']
            status = schema['openAPIV3Schema']['properties']['status']

            # ordered array of go structs as python dictionaries,
            # format {name:{field1:type, field2:type}}
            # ordered so that when rendered structs are declared before
            # first use.
            goStructs = []

            specType = parseYaml(spec, goStructs, '_spec')
            statusType = parseYaml(status, goStructs, '_status')
            specType = self.fixupTypeName(specType)
            statusType = self.fixupTypeName(statusType)

            # lines of text which are go statements defining the go structs
            defs = []
            for goStruct in goStructs:
                defs.extend(self.fixupGoStructs(goStruct))

            self.templateDict = {
                'defs': '\n'.join(defs),
                'crdType': crdTypeMap[self.key],
                'crdListType': '{}List'.format(crdTypeMap[self.key]),
                'specType': specType,
                'statusType': statusType,
            }

    def fixupTypeName(self, typname):
        if typname in self.typeNameMap.keys():
            typname = self.typeNameMap[typname]
        return typname

    def fixupGoStructs(self, goStruct):
        txt = []
        if len(goStruct.keys()) != 1:
            raise Exception('parsing error')
        for k, v in goStruct.items():
            if isinstance(v, dict):
                txt.append('type {} struct {{'.format(
                    self.fixupTypeName(k)))
                for nm, typ in sorted(v.items()):
                    if isinstance(typ, list):
                        txt.append('     {} []{} `json:"{}"`'.format(
                            capitalize(nm),
                            self.fixupTypeName(typ[0]),
                            nm)
                        )
                    else:
                        txt.append('     {} {} `json:"{}"`'.format(
                            capitalize(nm),
                            self.fixupTypeName(typ),
                            nm)
                        )
                txt.append('}')
            else:
                txt.append('type {} {} '.format(
                    self.fixupTypeName(k), v))
        txt.append('')
        return txt

    def generateGoFile(self):
        script_path = os.path.dirname(__file__)
        f = os.path.abspath(
            os.path.join(
                script_path,
                'api/types/v1alpha1/',
                '{}.go'.format(self.key)
                )
            )
        print('Generating file {}'.format(f))
        contents = template.format(**self.templateDict)
        with open(f, "w") as fp:
            fp.write(contents)
        os.system('gofmt -w {}'.format(f))


if __name__ == '__main__':
    for file in sys.argv[1:]:
        g = GenGoCrd(file)
        g.generateGoFile()
