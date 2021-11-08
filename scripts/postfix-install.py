#!/usr/bin/env python

import re
import yaml

reimg = re.compile(r'(?P<preceding>.*mayadata)/(?P<imagename>.*):(?P<tag>.*)')


def addDebug(comp_name, template_spec, container):
    '''
        - add -dev suffix to images
    '''
    gd = reimg.match(container['image']).groupdict()
    container['image'] = f'{gd["preceding"]}/{gd["imagename"]}-dev:{gd["tag"]}'
    print(f'Modifying {container["name"]} for debug')
    print(f'image: {container["image"]}')


def addCoverage(comp_name, template_spec, container):
    '''
        - add -cov suffix to images
        - add LLVM_PROFILE_FILE environment variable
        - add local volume for profile storage
        - add volumeMount to profile storage
    '''
    # FIXME: derive a name not present in volumeMounts and volumes
    cov_vol_name = 'coverage'
    cov_location = '/var/local/mayastor-coverage/'

    profraw = f'{comp_name}.profraw'
    env = {
        'name': 'LLVM_PROFILE_FILE',
        'value': f'{cov_location}{profraw}'
    }
    volumeMount = {
        'name': cov_vol_name,
        'mountPath': cov_location,
    }
    volume = {
        'name': cov_vol_name,
        'hostPath': {
            'path': cov_location,
            'type': 'DirectoryOrCreate',
        }
    }

    # FIXME: all sections below should check that the desired entries do not
    # exist before modification
    gd = reimg.match(container['image']).groupdict()
    imagename = f'{gd["preceding"]}/{gd["imagename"]}-cov:{gd["tag"]}'
    if imagename == container['image']:
        return False

    print(f'Modifying container "{container["name"]}" for coverage')

    print('change:\n'
          f'    image: {container["image"]}\n'
          'to:\n'
          f'    image: {imagename}'
          )
    container['image'] = imagename

    print(f'add:\n    env: {env}')
    try:
        container['env'].append(env)
    except KeyError:
        container['env'] = [env]

    print(f'add:\n    volumeMounts: {volumeMount}')
    try:
        container['volumeMounts'].append(volumeMount)
    except KeyError:
        container['volumeMounts'] = [volumeMount]

    print(f'add:\n    volumes: {volume}')
    try:
        template_spec['volumes'].append(volume)
    except KeyError:
        template_spec['volumes'] = [volume]

    return True


def process_yaml_file(yamlfile, **kwargs):
    '''
    process mayastor installation yaml file to meet requirements for one of
        - coverage
        - debug
    '''
    coverage = kwargs.get('coverage')
    debug = kwargs.get('debug')

    if not debug and not coverage:
        # release build, no post processing required
        return

    if debug and coverage:
        raise(Exception('Unsupported combination of debug and coverage'))

    changed = False
    with open(yamlfile, mode='r') as fp:
        contents = fp.read()
        ymls = list(yaml.safe_load_all(contents))
        for yml in ymls:
            comp_name = yml['metadata']['name']
            if 'spec' in yml and 'template' in yml['spec']:
                template_spec = yml['spec']['template']['spec']
                for container in template_spec['containers']:
                    # Only modify containers using mayastor images
                    m = reimg.match(container['image'])
                    if m is not None:
                        if coverage:
                            if addCoverage(comp_name, template_spec, container):
                                changed = True
                        if debug:
                            if addDebug(comp_name, template_spec, container):
                                changed = True

    if changed:
        # Only rewrite if contents changed
        with open(yamlfile, mode='w') as fp:
            _ = yaml.dump_all(ymls, fp)
        print(f'Updated {yamlfile}\n')


if __name__ == '__main__':
    from argparse import ArgumentParser
    description = r"""
    This python script amends mayastor install yaml files for
        - debug:
            - add appropriate suffix to image
        - coverage:
            - add appropriate suffix to image
            - add environment variable to generate coverage data
            - add local volume for storage of coverage data
    """
    print('postfix installation yaml files')
    parser = ArgumentParser(description=description)
    parser.add_argument('-d,--debug', dest='debug', action='store_true',
                        default=False,
                        help='Use development images'
                        )
    parser.add_argument('-c,--coverage', dest='coverage', action='store_true',
                        default=False,
                        help='Use coverage images'
                        )
    parser.add_argument('-b,--build_flags', dest='build_flags',
                        default=None,
                        help='build flags json'
                        )
    parser.add_argument('paths', nargs='*')
    _args = parser.parse_args()

    if _args.build_flags is not None:
        print(f'parsing build_flags: {_args.build_flags}')
        import json
        with open(_args.build_flags, 'r') as fp:
            bf = json.load(fp)
            for k,v in bf.items():
                if k in 'coverage':
                    _args.coverage = v
                if k in 'debug':
                    _args.debug = v
            _kwargs = bf

    if not _args.coverage and not _args.debug:
        print("nothing to do")
        exit(0)

    _kwargs = {
        k: v for k, v in vars(_args).items()
        if k not in ['paths', 'build_flags']
    }

    import os
    import traceback

    for spath in _args.paths:
        for (path, _, files) in os.walk(spath):
            files.sort()
            for filename in files:
                if filename.endswith('.yaml'):
                    filepath = os.path.join(path, filename)
                    print(filepath)
                    try:
                        process_yaml_file(filepath, **_kwargs)
                    except Exception:
                        traceback.print_exc()
