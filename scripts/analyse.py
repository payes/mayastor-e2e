#!/usr/bin/env python
import re
import sys
import os

# match lines like
# k8stest.InitTesting(t, "Basic volume IO tests, NVMe-oF TCP and iSCSI", "basic_volume_io")
re_InitTesting = re.compile(
    r'^\s*k8stest.InitTesting\((?P<t>.*)\s*,\s*"(?P<classname>.*)"\s*,\s*(?P<reportname>.*)\s*\)')

# match lines like
# var _ = Describe("Mayastor Volume IO test", func() {
re_Describe = re.compile(
    r'^.*=(\s*|\s*ginkgo\.)Describe\s*\(\s*"(?P<desc>.*)"\s*,\s*func\s*\(\s*\)\s*{')

# match lines like
# It("should verify an NVMe-oF TCP volume can process IO on a Filesystem volume with immediate binding", func() {
re_It = re.compile(
    r'^(\s*|\s*ginkgo\.)It\s*\(\s*"(?P<it>.*)"\s*,\s*func\s*\(\s*\)\s*{')

# match It clauses which are not matched by re_It,
# where we cannot derive the actual value of the string,
# for example:
# ginkgo.It(fmt.Sprintf("should delete PV with reclaimPolicy %q [mayastor-csi.openebs.io]", v1.PersistentVolumeReclaimDelete), func() {
re_ItPromiscuous = re.compile(
    r'^(\s*|\s*ginkgo\.)It\s*\(\s*(?P<it>.*),\s*func\s*\(\s*\)\s*{')

# match lines like
# ginkgo.RunSpecsWithDefaultAndCustomReporters(t, "CSI E2E Suite", rep.GetReporters("csi"))
re_RunSpecs = re.compile(
    r'^\s*ginkgo.RunSpecsWithDefaultAndCustomReporters\s*\(\s*\w\s*,\s*"(?P<classname>.*)"\s*,')


def scrapeSourceFileClassnames(f, classnames, verbose=False):
    """
    Examine file contents to extract the test classname
    either from InitTesting call or RunSpecs* calls
    and populate the classnames map
    """
    with open(f, 'r') as fp:
        lines = [x.strip() for x in fp.readlines()]
    classname = None
    for line in lines:
        m = re_InitTesting.match(line)
        if m is not None:
            classname = m.groupdict()['classname']
        else:
            # 3rd party source
            m = re_RunSpecs.match(line)
            if m is not None:
                classname = m.groupdict()['classname']

    if classname is None:
        return

    dirname = os.path.dirname(f)
    if dirname in classnames:
        raise(
            Exception(f'{dirname} already has a classname',
                      f' {classnames[dirname]}'))
    classnames[dirname] = classname


def scrapeSourceFile(f, d2s, classname, verbose=False):
    """
    Examine source files for Describe, and It clauses
    and populate the map of xray test definitions to source files
    """

    warnings = []
    if classname is None:
        return

    with open(f, 'r') as fp:
        lines = [x.strip() for x in fp.readlines()]

    desc = None
    for ix in range(len(lines)):
        line = lines[ix]
        line_num = ix + 1
        m = re_Describe.match(line)
        if m is not None:
            desc = m.groupdict()['desc']

        m = re_It.match(line)
        if m is not None:
            it = m.groupdict()['it']
            definition = '{}.{} {}'.format(classname, desc, it)
            definition = definition.replace('\\', '')
            if definition in d2s.keys():
                org = d2s[definition]
                print('\n'
                      f'WARNING!!!! duplicate definitions {definition}\n'
                      f'{org["file"]}:{org["line_num"]}\n'
                      f'{org["line"]}\n'
                      f'{f}:{line_num}:\n'
                      f'{line}\n',
                      file=sys.stderr)
                warnings.append(
                    'WARNING!!!!!! duplicate definition'
                    f'{definition} {f} it:{it}'
                )
            else:
                d2s[definition] = {'file': f,
                                   'line_num': line_num, 'line': line}
        else:
            m = re_ItPromiscuous.match(line)
            if m is not None:
                print('\n'
                      'WARNING!!!!! unhandled it clause\n'
                      f'{f}:{line_num}\n{line}',
                      file=sys.stderr)
                warnings.append(
                    f'WARNING!!!!! unhandled it clause\n{f}:{line_num}\n{line}'
                )

    return warnings


def scrapeSources(srcpath):
    import os
    srcpath = os.path.realpath(srcpath)
    dirs = [x.name for x in os.scandir(srcpath) if x.is_dir()]

    # map of test definitions to source file
    defs2src = {}
    # map of test classnames to source directory
    classnames = {}
    for dir in dirs:
        testpath = os.path.join(srcpath, dir)
        gofiles = [x.name for x in os.scandir(
            testpath) if x.is_file() and x.name.endswith('.go')]
        for goFile in gofiles:
            scrapeSourceFileClassnames(
                os.path.join(testpath, goFile), classnames)
        for goFile in gofiles:
            _ = scrapeSourceFile(os.path.join(testpath, goFile),
                                 defs2src, classnames.get(testpath))

    return {k: v['file'] for k, v in defs2src.items()}
