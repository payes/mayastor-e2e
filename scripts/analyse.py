#!/usr/bin/env python
from datetime import datetime
import json
import re
import sys
import os

re_InitTesting = re.compile(
    r'^\s*k8stest.InitTesting\((?P<t>.*)\s*,\s*"(?P<classname>.*)"\s*,\s*(?P<reportname>.*)\s*\)')
re_Describe = re.compile(
    r'^.*=(\s*|\s*ginkgo\.)Describe\s*\(\s*"(?P<desc>.*)"\s*,\s*func\s*\(\s*\)\s*{')
re_It = re.compile(
    r'^(\s*|\s*ginkgo\.)It\s*\(\s*"(?P<it>.*)"\s*,\s*func\s*\(\s*\)\s*{')
re_ItPromiscuous = re.compile(
    r'^(\s*|\s*ginkgo\.)It\s*\(\s*(?P<it>.*),\s*func\s*\(\s*\)\s*{')
re_RunSpecs = re.compile(
    r'^\s*ginkgo.RunSpecsWithDefaultAndCustomReporters\s*\(\s*\w\s*,\s*"(?P<classname>.*)"\s*,')


def scrapeSourceFileClassnames(f, classnames, verbose=False):
    """ it should be obvious from the source
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
    """ it should be obvious from the source
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
                    f'WARNING!!!!!! duplicate definition {definition} {f} it:{it}'
                )
            else:
                d2s[definition] = {'file': f,
                                   'line_num': line_num, 'line': line}
        else:
            m = re_ItPromiscuous.match(line)
            if m is not None:
                print('\n'
                      f'WARNING!!!!! unhandled it clause\n{f}:{line_num}\n{line}',
                      file=sys.stderr)
                warnings.append(
                    f'WARNING!!!!! unhandled it clause\n{f}:{line_num}\n{line}'
                )

    return warnings


def scrapeSources(srcpath):
    import os
    srcpath = os.path.realpath(srcpath)
    dirs = [x.name for x in os.scandir(srcpath) if x.is_dir()]

    defs2src = {}
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


# def tidyScrapeSourceFileClassnames(f, classnames, verbose=False):
#    """ it should be obvious from the source
#    """
#
#    def tidy_the_line(txt, org, tidy):
#        tidy_line = txt.replace(org, tidy, 1)
#        if tidy_line != txt:
#            if verbose:
#                print('>>>', txt, file=sys.stderr)
#                print('<<<', tidy_line, file=sys.stderr)
#        return tidy_line
#
#    with open(f, 'r') as fp:
#        lines = [x.strip() for x in fp.readlines()]
#    classname = None
#    for line in lines:
#        m = re_InitTesting.match(line)
#        if m is not None:
#            classname = m.groupdict()['classname']
#            tidy_classname = classname.strip()
#            if classname != tidy_classname:
#                # TODO change the line in the array
#                tidy_the_line(line, classname, tidy_classname)
#        else:
#            # 3rd party source
#            m = re_RunSpecs.match(line)
#            if m is not None:
#                classname = m.groupdict()['classname']
#
#    if classname is None:
#        return
#    dirname = os.path.dirname(f)
#    if dirname in classnames:
#        raise(
#            Exception(f'{dirname} already has a classname'
#                      f' {classnames[dirname]}'))
#    classnames[dirname] = classname
#
#
# def tidyScrapeSourceFile(f, d2s, clazzname, verbose=False):
#    """ it should be obvious from the source
#    """
#
#    def tidy_the_line(txt, org, tidy):
#        tidy_line = txt.replace(org, tidy, 1)
#        if tidy_line != txt:
#            if verbose:
#                print('>>>', txt, file=sys.stderr)
#                print('<<<', tidy_line, file=sys.stderr)
#        return tidy_line
#
#    with open(f, 'r') as fp:
#        lines = [x.strip() for x in fp.readlines()]
#    classname = None
#    for line in lines:
#        m = re_InitTesting.match(line)
#        if m is not None:
#            classname = m.groupdict()['classname']
#            tidy_classname = classname.strip()
#            if classname != tidy_classname:
#                # TODO change the line in the array
#                tidy_the_line(line, classname, tidy_classname)
#        else:
#            # 3rd party source
#            m = re_RunSpecs.match(line)
#            if m is not None:
#                classname = m.groupdict()['classname']
#
#    if classname is None:
#        classname = clazzname
#        tidy_classname = clazzname
#
#    if classname is None and clazzname is None:
#        return
#
#    desc = None
#    for ix in range(len(lines)):
#        line = lines[ix]
#        m = re_Describe.match(line)
#        if m is not None:
#            desc = m.groupdict()['desc']
#            tidy_desc = desc.strip()
#            if tidy_desc[-1] != ':':
#                tidy_desc += ':'
#                if desc != tidy_desc:
#                    # TODO change the line in the array
#                    tidy_the_line(line, desc, tidy_desc)
#
#        m = re_It.match(line)
#        if m is not None:
#            it = m.groupdict()['it']
#            definition = '{}.{} {}'.format(classname, desc, it)
#            tidy_it = it.strip()
#            if it != tidy_it:
#                # TODO change the line in the array
#                tidy_the_line(line, it, tidy_it)
#                pass
#            tidy_definition = '{}.{} {}'.format(
#                tidy_classname, tidy_desc, tidy_it)
#            if verbose:
#                print('---', definition, file=sys.stderr)
#                if definition != tidy_definition:
#                    print('+++', tidy_definition, file=sys.stderr)
#            definition = definition.replace('\\', '')
#            if definition in d2s.keys():
#                print(f'WARNING!!!!!! duplicate definition {definition}',
#                      file=sys.stderr)
#            else:
#                d2s[definition] = f


def testRunComprehensions(runs, verbose=False):
    """
    rejig the data in the list of test runs for easier human comprehension
    The primary keys are JIRA Key and the test definition test,
    so rekey the data on those values.
    We are primarily interested in
        - status of the test run,
        - sorting the status based on the timestamp.
    """
    from datetime import datetime

    tmqs = {}
    tuns = {}
    tdefs = {}

    for run in runs:
        jiraKey = run["test"]["jira"]["key"]
        ts = run['finishedOn']
        status = run["status"]["name"]
        uns = run["unstructured"]

        # TODO remove this check, when we do the nomenclature fixes
        # unstructured will change for the test
        # in the JIRA Web Pages for a Test "unstructured" is the definition
        # field in test details.
        if jiraKey in tuns:
            if uns != tuns[jiraKey]:
                raise(
                    Exception(f'{tuns[jiraKey]} != {uns} for {run}'))
        else:
            tuns[jiraKey] = uns

        if jiraKey not in tmqs:
            tmqs[jiraKey] = {
                ts: [status]
            }
        else:
            if ts in tmqs[jiraKey]:
                tmqs[jiraKey][ts].extend(status)
            else:
                tmqs[jiraKey][ts] = [status]

        if uns not in tdefs:
            tdefs[uns] = {
                ts: [status]
            }
        else:
            if ts in tdefs[uns]:
                tdefs[uns][ts].extend(status)
            else:
                tdefs[uns][ts] = [status]

    if verbose:
        for mq, v in tmqs.items():
            print(mq, tuns[mq])
            for ts in sorted(v, reverse=True):
                dt = datetime.fromtimestamp(float(ts)/1000)
                print('\t', ts, dt.ctime(), v[ts])

        print('--------------------------------')

        for tdef, v in tdefs.items():
            print(tdef)
            for ts in sorted(v, reverse=True):
                dt = datetime.fromtimestamp(float(ts)/1000)
                print('\t', ts, dt.ctime(), v[ts])

    return tmqs, tdefs, tuns


def printDefRuns(defRuns):
    for k, v in defRuns.items():
        print(k)
        for ts in sorted(v, reverse=True):
            dt = datetime.fromtimestamp(float(ts)/1000)
            print('\t', ts, dt.ctime(), v[ts])


def gradeDefRuns(defRuns):

    grades = {}

    def countPasses(v):
        """ Count number of PASSED until
            - all PASSED
            - !PASSED
            returns count and boolean allPasses
        """
        count = 0
        for stat in v:
            if stat == 'PASSED':
                count += 1
            else:
                return count, False
        return count, True

    for k, v in defRuns.items():
        latestConsecutivePassCount = 0
        for ts in sorted(v, reverse=True):
            c, allPasses = countPasses(v[ts])
            latestConsecutivePassCount += c
            if not allPasses:
                break
        grades[k] = latestConsecutivePassCount

    return grades


def gradedSrcRuns(gradedDefs, defs2src):
    srcGrades = {}
    for k, v in gradedDefs.items():
        try:
            src = defs2src[k]
        except KeyError:
            print("ignoring ", k, file=sys.stderr)
            continue
        sk = src.split('/')[-2]
        try:
            srcGrades[sk].append({k: v})
        except KeyError:
            srcGrades[sk] = [{k: v}]

    return srcGrades


def printGradedSrcRuns(srcGrades, dump=False):
    if dump:
        print(json.dumps(srcGrades, sort_keys=True, indent=4))
    for k, v in srcGrades.items():
        grades = [e[ek] for e in v for ek in e]
        print(min(grades), k)


if __name__ == '__main__':

    #    with open(sys.argv[1]) as f:
    #        runz = json.load(f)
    #        mqRuns, defRuns = testRunComprehensions(runz, False)
    #
    #    with open('/tmp/defRuns.json', 'w') as fp:
    #        json.dump(defRuns, fp, sort_keys=True, indent=4)
    #
    #    # printDefRuns(defRuns)
    #    gradedDefs = gradeDefRuns(defRuns)
    #
    #    with open(sys.argv[2]) as f:
    #        ds = json.load(f)
    #
    #    defs2src = ds['defs']
    #    printGradedSrcRuns(gradedSrcRuns(gradedDefs, defs2src))
    if len(sys.argv) > 1:
        defs2src = scrapeSources(os.path.realpath(sys.argv[1] + '/src'))
    else:
        defs2src = scrapeSources(os.path.realpath(sys.path[0] + '/../src'))
