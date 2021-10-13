#!/usr/bin/env python
import os
import re
import shutil
import sys
import yaml

verbose = True


def adjustSpec(src_specyaml, dst_specyaml):
    '''
    This function adds additionalProperties=True to the JsonGeneric schema.
    This is a workaround to facilitate generation of openapi go client code
    when using openapi-generator-cli (5.2.1), which fails to generate code
    without this clause.
    '''

    print(f'Patching {src_specyaml} and saving to {dst_specyaml}')
    with open(src_specyaml, mode='r') as fp:
        contents = fp.read()
        spec = yaml.safe_load(contents)

    try:
        ap = spec['components']['schemas']['JsonGeneric']['additionalProperties']
        print("Found spec['components']['schemas']['JsonGeneric']['additionalProperties'] = ",
              ap)
    except KeyError:
        print(
            "Adding spec['components']['schemas']['JsonGeneric']['additionalProperties'] = True")
        spec['components']['schemas']['JsonGeneric']['additionalProperties'] = True

    with open(dst_specyaml, mode='w') as fp:
        spec = yaml.safe_dump(spec, fp)


# support functions to fix the generated code
# most fixes are trivial except for fixupOneOf
def read_file(filename):
    try:
        with open(filename) as fp:
            curr_lines = [x.rstrip() for x in fp.readlines()]
    except Exception as e:
        print(e, file=sys.stderr)
        curr_lines = []
    return curr_lines


def write_file(filename, lines):
    try:
        with open(filename, "w") as fp:
            for line in lines:
                fp.write(line)
                fp.write('\n')
    except Exception as e:
        print(e, file=sys.stderr)


def fixupOneOf(filename, lines):
    '''
    Replace incorrect generated code with code that compiles.
    '''
# < 	interface{} *interface{}
# ---
# > 	ifc *interface{}
    reIfcOneOf = re.compile(
        r'(?P<lsp>\s+)(?P<i>interface{})(?P<sep1>\s+)\*interface{}\s*$')

# < func interface{}As...
# ---
# > func ifcAs...
    reIfcFuncAs = re.compile(r'(?P<lsp>\s*)func\s+interface\{\}As(?P<rest>.*)')

# < {interface{}:...
# ---
# > {ifc:...
    refIfcDecl = re.compile(r'(?P<left>.*)interface{}:(?P<rest>.*)')

# < .interface{}
# ---
# > .ifc
    reIfcMember = re.compile(r'(?P<left>.*)\.interface{}(?P<right>.*)')

# < jsoninterface{}
# ---
# > jsonifc
    reJsonIfc = re.compile(r'(?P<left>.*)jsoninterface\{\}(?P<right>.*)')

    changed = False
    for ix in range(0, len(lines)):
        line = lines[ix]
        if line.startswith('//'):
            continue
        m = reIfcOneOf.match(line)
        if m is not None:
            gd = m.groupdict()
            line = gd['lsp'] + 'ifc' + gd['sep1'] + '*interface{}'
            if verbose:
                if not changed:
                    print(f'+++ b{filename}')
                print(f'{ix}c{ix}')
                print("<", lines[ix])
                print('---')
                print('>', line)
                print('')
            lines[ix] = line
            changed = True
            continue
        if not changed:
            continue
        m = reIfcFuncAs.match(line)
        if m is not None:
            gd = m.groupdict()
            line = gd['lsp'] + 'func ifcAs' + gd['rest']
            if verbose:
                print(f'{ix}c{ix}')
                print("<", lines[ix])
                print('---')
                print('>', line)
                print('')
            lines[ix] = line
            changed = True
            continue
        m = refIfcDecl.match(line)
        if m is not None:
            gd = m.groupdict()
            line = gd['left'] + 'ifc:' + gd['rest']
            if verbose:
                print(f'{ix}c{ix}')
                print("<", lines[ix])
                print('---')
                print('>', line)
                print('')
            lines[ix] = line
            changed = True
            continue
        m = reJsonIfc.match(line)
        if m is not None:
            gd = m.groupdict()
            line = gd['left'] + 'jsonifc' + gd['right']
            if verbose:
                print(f'{ix}c{ix}')
                print("<", lines[ix])
                print('---')
                print('>', line)
                print('')
            lines[ix] = line
            changed = True
        m = reIfcMember.match(line)
        if m is not None:
            gd = m.groupdict()
            line = gd['left'] + '.ifc' + gd['right']
            if verbose:
                print(f'{ix}c{ix}')
                print("<", lines[ix])
                print('---')
                print('>', line)
                print('')
            lines[ix] = line
            changed = True
            continue
    return True, lines


def fixupXX(filename, lines):
    '''
    Replace incorrect generated code for response range
    '''
    re_Nxx = re.compile(
        r'(?P<lsp>\s*if\s+)(?P<exp>.*)\s+==\s+(?P<val>\d)XX(?P<rest>.*)')
# < 	(...) == 4XX {
# ---
# > 	(...) >= 400 && (...) <= 499 {

# < 	(...) == 5XX {
# ---
# > 	(...) >= 500 && (...) <= 599 {
    changed = False
    for ix in range(0, len(lines)):
        line = lines[ix]
        m = re_Nxx.match(line)
        if m is not None:
            gd = m.groupdict()
            line = gd['lsp'] + gd['exp'] + " >= " + gd['val'] + "00 " + \
                "&& " + gd['exp'] + " <= " + gd['val'] + "99" + gd['rest']
            if verbose:
                if not changed:
                    print(f'+++ b{filename}')
                print(f'{ix}c{ix}')
                print("<", lines[ix])
                print('---')
                print('>', line)
                print('')
            lines[ix] = line
            changed = True
    return changed, lines


def fixup500(filename, lines):
    '''
    Replace incorrect generated code for response range
    '''
    changed = False
    re_500 = re.compile(
        r'(?P<lsp>\s*if\s+)(?P<mid>localVarHTTPResponse.StatusCode)\s+>=\s+500$')
    for ix in range(0, len(lines)):
        line = lines[ix]
        m = re_500.match(line)
        if m is not None:
            gd = m.groupdict()
            line = gd['lsp'] + gd['mid'] + " >= 500 {"
            if verbose:
                if not changed:
                    print(f'+++ b{filename}')
                print(f'{ix}c{ix}')
                print("<", lines[ix])
                print('---')
                print('>', line)
                print('')
            lines[ix] = line
            changed = True

    return changed, lines


if __name__ == '__main__':
    scriptdir = os.path.abspath(os.path.join(os.path.dirname(__file__)))
    repodir = os.path.abspath(os.path.join(scriptdir, '..'))
    artifactsdir = os.path.join(repodir, 'artifacts')
    openapidir = os.path.join(artifactsdir, 'openapi')
    gendir = os.path.join(repodir, 'src/generated/openapi')

    from argparse import ArgumentParser
    parser = ArgumentParser()

    parser.add_argument('spec', default=None,
                        help="path to control plane openapi spec yaml file"
                        )
    parser.add_argument(
        '--specout', default=os.path.join(openapidir, 'spec.yaml'),
        help='path to "fixed" spec yaml file'
    )
    parser.add_argument('--speconly', action="store_true", default=False,
                        help='only "fix" spec'
                        )
    parser.add_argument('--dest', default=openapidir,
                        help='destination for generate go client source'
                        )

    args = parser.parse_args()

    if args.speconly:
        adjustSpec(args.spec, args.specout)
    else:
        print(f'###########################################################')
        print(f'Generating openapi go client from {args.spec} into {gendir}')
        try:
            shutil.rmtree(openapidir)
        except FileNotFoundError:
            pass
        os.makedirs(openapidir)

        adjustSpec(args.spec, args.specout)
        jarfile = os.path.join(scriptdir, 'openapi-generator-cli-5.2.1.jar')
        addnprops = 'enumClassPrefix=true,structPrefix=true,disallowAdditionalPropertiesIfNotPresent=false'

        cmd = f'java -jar {jarfile} generate -i {args.specout} -g go --additional-properties={addnprops} -o {args.dest}'
        print(f'Running:{cmd}')
        sys.stdout.flush()
        os.system(cmd)

        try:
            print(f'Deleting all files in {gendir}')
            shutil.rmtree(gendir)
        except FileNotFoundError:
            pass

        print(f'Copying generated code from {openapidir} to {gendir}')
        shutil.copytree(openapidir, gendir)
        # remove go module files.
        try:
            os.remove(os.path.join(gendir, 'go.mod'))
        except FileNotFoundError:
            pass
        try:
            os.remove(os.path.join(gendir, 'go.sum'))
        except FileNotFoundError:
            pass

        gofiles = [os.path.join(gendir, f)
                   for f in os.listdir(args.dest) if f.endswith('.go')]
        print(f'Fixing up  generated code in {gendir}')
        for file in gofiles:
            lines = read_file(file)
            saveFile = False
            changed, lines = fixupXX(file, lines)
            saveFile = saveFile or changed
            changed, lines = fixup500(file, lines)
            saveFile = saveFile or changed
            changed, lines = fixupOneOf(file, lines)
            saveFile = saveFile or changed
            if saveFile:
                write_file(file, lines)
