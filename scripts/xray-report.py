#!/usr/bin/env python3
from datetime import datetime
import html
import json
import sys
from time import sleep
import traceback
import xraygraphql
import analyse

htmlStart = '''
<!DOCTYPE html>

<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">

  <title>Mayastor E2E Results</title>

  <style>
  table, th {
  border: 2px solid white;
  border-collapse: collapse;
  }
  th {
  background-color: #96D4D4;
  font-weight: normal;
  padding: 3px;
  }
  td {
  border: 1px solid white;
  border-collapse: separate;
  border-radius: 5px;
  }
  a {
  text-decoration: none;
  }
  body {
    background-color: #d4d4d4;
  }
  </style>
</head>
<body>
<h1>Mayastor E2E Results</h1>
<meta name="description" content="Mayastor E2E test results.">
<meta name="author" content="xray-report.py">
'''

htmlEnd = '''
</body>
</html>
'''


def jsonSave(filename, data):
    with open(filename, 'w') as fp:
        json.dump(data, fp, sort_keys=True, indent=4)


def jsonUpdate(filename, data):
    '''
    Dumb update - works for dictionaries only
    '''
    existing = {}
    try:
        with open(filename, 'r') as fp:
            existing = json.load(fp)
    except FileNotFoundError:
        print(f'New {filename}', file=sys.stderr)
        pass

    existing.update(data)

    with open(filename, 'w') as fp:
        json.dump(data, fp, sort_keys=True, indent=4)


def jsonExists(filename):
    '''
    Slightly more the os.exists
    '''
    try:
        with open(filename, 'r') as fp:
            _ = json.load(fp)
            return True
    except FileNotFoundError:
        pass

    return False


def jsonLoad(filename):
    '''
    Dumb update - works for dictionaries only
    '''
    try:
        with open(filename, 'r') as fp:
            data = json.load(fp)
            return data
    except FileNotFoundError:
        pass

    return None


class tableCell():
    def __init__(self, txt='', link=None, titl=None, colspan=0, style=None,
                 ts=0):
        self.txt = txt
        self.link = link
        if titl is not None:
            titl.replace('"', '&quot;')
            titl.replace('<', '&lt;')
            titl.replace('>', '&gt;')
            titl = html.escape(titl)
        self.titl = titl
        self.colspan = colspan
        self.celltype = 'td'
        self.style = style
        # ts is not used for display
        self.ts = ts

    def __str__(self):
        return f'{self.txt} {self.ts}'

    def render(self):
        f = []
        f.append(f'<{self.celltype}')

        if self.titl:
            f.append(f' title="{self.titl}"')

        if self.colspan != 0:
            f.append(f' colspan="{self.colspan}"')

        if self.style:
            f.append(f' style="{self.style}"')

        f.append('>')

        if self.link:
            f.append(f'<a href="{self.link}" target="_blank">{self.txt}</a>')
        else:
            f.append(f'{self.txt}')

        f.append(f'</{self.celltype}>')

        return ''.join(f)


class testRunCell(tableCell):
    def __init__(self, txt='', link=None, titl=None):
        # ballot box check #x2611
        # ballot box x #x2612
        # ballot box  #x2610
        # check #x2713
        # ballot X #x2716
        self.failed = False
        if txt == 'P':
            style = (
                'text-align:center;'
                ' vertical-align:middle;'
                ' background-color:#93E9BE;'
            )
            # Invisible, but clickable
            txt = '<span style="opacity:0;">&#x2713;</span>'
        elif txt == 'F':
            style = (
                'text-align:center;'
                ' vertical-align:middle;'
                ' background-color:#FFD1DC;'
            )
            # Invisible, but clickable
            txt = '<span style="opacity:0;">&#x2717;</span>'
            self.failed = True
        else:
            style = (
                'text-align:center;'
                ' vertical-align:middle;'
            )
            txt = '<span style="opacity:0;"> </span>'

        tableCell.__init__(self, txt=txt, link=link, titl=titl, style=style)


class headerCell(tableCell):
    def __init__(self, txt='', link=None, titl=None, colspan=0):
        tableCell.__init__(
            self, txt=txt, link=link, titl=titl, colspan=colspan,
        )
        self.celltype = 'th'

    def __str__(self):
        return f'{self.txt} {self.colspan}'


class testCell(tableCell):
    def __init__(self, txt='', link=None, titl=None):
        tableCell.__init__(
            self, txt=txt, link=link, titl=titl,
            style='background-color:#96d4d4;'
        )


class srcCell(tableCell):
    def __init__(self, txt=''):
        tableCell.__init__(
            self, txt=txt,
            style='font-size:small;'
        )


class plainCell(tableCell):
    def __init__(self, txt='', link=None, titl=None):
        tableCell.__init__(
            self, txt=txt, link=link, titl=titl,
            style='text-align:center; vertical-align:middle'
        )


def issueLink(issueKey):
    return f'https://mayadata.atlassian.net/browse/{issueKey}'


def runLink(execKey, issueKey):
    # https://mayadata.atlassian.net/plugins/servlet/ac/com.xpandit.plugins.xray/execution-page?ac.testExecIssueKey=MQ-2532&ac.testIssueKey=MQ-2425
    return ('https://mayadata.atlassian.net/plugins/servlet/ac/'
            'com.xpandit.plugins.xray/execution-page?'
            f'ac.testExecIssueKey={execKey}'
            f'&amp;ac.testIssueKey={issueKey}')


def orderTests(tests):
    """
    order the tests on numerical value - so that we generate tables
    which are ordered in a manner which is easy to comprehend,
    for example MQ-1 and MQ-9 will precede MQ-10
    """
    kiis = sorted(
        tests.keys(),
        key=lambda test: int(test.split('-')[-1])
    )

    ordered = []
    for k in kiis:
        if k in ordered:
            continue
        ordered.append(k)
        def1 = tests[k]['unstructured'].split('.')[0]
        for kk in kiis:
            if kk in ordered:
                continue
            if def1 == tests[kk]['unstructured'].split('.')[0]:
                ordered.append(kk)
    return ordered


def writeTable(fp, hdrs, table, h2=None):

    if h2 is not None:
        print(f'<p><h2>{h2}</h2>', file=fp)
        writeTable(fp, hdrs, table, None)
        print('</p><br><hr>', file=fp)
        return

    def writeRows(rows):
        for row in rows:
            print('<tr>', file=fp)
            print(row[0].render(), file=fp)
            for col in row[1:]:
                print(col.render(), file=fp)
            print('</tr>', file=fp)

    if len(table) == 0:
        return

    print('<table>', file=fp)
    writeRows(hdrs)
    writeRows(table)
    print('</table>', file=fp)


def startReport(fp, annotation='', testEnv=None):
    if testEnv:
        te = f'({", ".join(testEnv)})'
    else:
        te = ''
    hs = htmlStart.replace('Mayastor E2E Results',
                           f'Mayastor E2E Results {annotation} {te}'.strip())
    print(hs, file=fp)
    tplink = issueLink(_args.jiraKey)
    print(f'<h2>Test Plan <a href="{tplink}">{_args.jiraKey}</a></h2>',
          file=fp)
    if testEnv:
        print(f'<p>Test Environments: {testEnv}</p>',
              file=fp)


def endReport(fp):
    print(htmlEnd, file=fp)


def main():

    defs2src = analyse.scrapeSources(
        os.path.realpath(sys.path[0] + '/../src'))
    defs2src = {
        k: {
            'location': v,
            'classname': k.split('.')[0],
        } for k, v in defs2src.items()
    }

    tests = jsonLoad(testListFile)
    for k, test in tests.items():
        if test['unstructured'] is None:
            print(f'Warning set definition to unknown for {k}')
            test['unstructured'] = 'Unknown.unknown'
    testKeys = orderTests(tests)

    executions = [v for _, v in jsonLoad(testExecsFile).items()]

    # reverse for reverse chronological order
    # first attempt used 'lastModified', but that gives spurious
    # ordering, and there is no equivalent of 'createdTime'
    # so we order on MQ-number
    executions = sorted(executions,
                        key=lambda d: int(d['jira']['key'].split('-')[-1]),
                        reverse=True)

    testEnvsCombis = []

    defaultTestEnv = _args.testenvs.split(',')
    # Work out the timestamp for each execution
    # Unfortunately there isn't a field we can use to derive the
    # timestamp for the test execution. So we record the earliest
    # timestamp of a test run in the test execution, and use that
    # to generate the date information for the execution
    for exe in executions:
        if exe.get('testEnvironments') == []:
            exe['testEnvironments'] = defaultTestEnv

        exe['testEnvironments'] = sorted(exe['testEnvironments'])

        execDict = {
            d['test']['jira']['key']: d for d in jsonLoad(
                f'{_args.cachedir}/exec.{exe["jira"]["key"]}.json')
        }
        timestamp = 0
        for testKey in testKeys:
            if testKey in execDict:
                test = execDict[testKey]
                if test['finishedOn'] is None:
                    test['finishedOn'] = 0
                tmp = test['finishedOn']
                try:
                    finished = int(tmp)/1000
                except ValueError:
                    finished = int(datetime.strptime(
                        tmp, "%Y-%m-%dT%H:%M:%S.%fZ").timestamp())
                if timestamp == 0 or finished < timestamp:
                    timestamp = finished
        exe['timestamp'] = timestamp
        if timestamp == 0:
            print(f'No test runs found, will discard test execution {exe}',
                  file=sys.stderr)

    executions = [exe for exe in executions if exe['timestamp'] != 0]

    for exe in executions:
        if exe['testEnvironments'] not in testEnvsCombis:
            testEnvsCombis.append(exe['testEnvironments'])

    for testEnvironments in testEnvsCombis:
        createReport(
            [
                exe for exe in executions
                if testEnvironments == exe['testEnvironments']
            ],
            tests,
            testKeys,
            defs2src,
            testEnvironments
        )

    createReport(
        executions[:],
        tests,
        testKeys,
        defs2src,
        ['all']
    )


def createReport(executions, tests, testKeys, defs2src, testEnvironments):
    if len(executions) > _args.columns:
        executions = executions[:_args.columns]

    def envString(te):
        return '\n\t'.join(te)

    execRow = [
        headerCell(
            titl=(
                f'{d["jira"]["summary"]}\ntestEnvironments:'
                f'\n\t{envString(d["testEnvironments"])}'
            ),
            link=issueLink(d['jira']['key']),
            txt=datetime.fromtimestamp(float(d['timestamp'])).strftime('%d')
        )
        for d in executions
    ]
    execRow.insert(0, headerCell(txt='Test\\Day'))
    execRow.append(headerCell(txt='Location'))

    # Create header rows for year,month, and day ordered by the
    # the list of executions
    hdrRows = [
        [headerCell(txt='Year')],
        [headerCell(txt='Month')],
        execRow
    ]
    for exe in executions:
        dt = datetime.fromtimestamp(float(exe['timestamp']))
        dt.replace(second=0, microsecond=0)
        year = dt.strftime('%Y')
        if year == hdrRows[0][-1].txt:
            hdrRows[0][-1].colspan += 1
        else:
            hdrRows[0].append(headerCell(txt=year, colspan=1))

        mon = dt.strftime('%b')
        if mon == hdrRows[1][-1].txt:
            hdrRows[1][-1].colspan += 1
        else:
            hdrRows[1].append(headerCell(txt=mon, colspan=1))

    # The objective is to create a 2 dimensional array from which rendering
    # as an html table is a prettifying but dumb op.
    # This means filling missing test runs with placeholders which render as
    # empty table cells

    blank = tableCell()
    blanks = [blank for _ in executions]
    # for source column
    blanks.append(blank)

    # Create an "empty" test table
    # add the rows for each test:
    #   column 0 is the test
    #   the rest are blank
    atable = []
    for testKey in testKeys:
        test = tests[testKey]
        atable.append(
            [testCell(
                txt=testKey,
                titl=test['unstructured'],
                link=issueLink(testKey))
             ])
        atable[-1].extend(blanks)
        src_info = defs2src.get(test['unstructured'])
        if src_info is not None:
            atable[-1][-1] = srcCell(
                txt=src_info['location'].split('/')[-2]
            )

    # populate the test results for test executions
    # in the test table
    col = 1     # col 0 is the test header cell
    for exe in executions:
        execDict = {
            d['test']['jira']['key']: d for d in jsonLoad(
                f'{_args.cachedir}/exec.{exe["jira"]["key"]}.json')
        }
        row = 0
        xtime = datetime.fromtimestamp(float(exe['timestamp'])).ctime()
        for testKey in testKeys:
            if testKey in execDict:
                test = execDict[testKey]
                atable[row][col] = testRunCell(
                    txt=test['status']['name'][0],
                    titl=f'{xtime}\n\n{test["results"][0]["log"]}',
                    link=runLink(exe["jira"]["key"], testKey),
                )
            row += 1
        col += 1

    # order the rows based on the age and freuency of failures
    score_sample_depth = 14
    scores = {}
    for ix in [x for x in range(len(atable))
               if not atable[x][0].titl.split('.')[-1] in
               ['BeforeSuite', 'AfterSuite']
               ]:
        tmp = atable[ix][1:score_sample_depth]
        weight = 2 ** 32
        score = 0
        for cell in tmp:
            if isinstance(cell, testRunCell) and not cell.failed:
                score += weight
            weight /= 2

        weight = len(atable[ix])
        for cell in atable[ix][1:]:
            if isinstance(cell, testRunCell) and not cell.failed:
                score += weight
                weight -= 1
        scores[ix] = score

    # gtable is the table of tests ordered on score i.e. age and freuency of
    # failures.
    gtable = [atable[ix]
              for ix in sorted(scores.keys(), key=lambda sc: scores[sc])]

    # ttable is the table of tests excluding BeforeSuite, AfterSuite
    ttable = [atable[ix] for ix in range(len(atable))
              if not atable[ix][0].titl.split('.')[-1] in
              ['BeforeSuite', 'AfterSuite']
              ]

    # ftable is the table of tests that have failed in the latest iteration
    ftable = [atable[ix] for ix in range(len(atable))
              if isinstance(atable[ix][1], testRunCell)
              and atable[ix][1].failed]

    # ntable is the table of tests that have not run in the latest iteration
    # excluding BeforeSuite, AfterSuite
    ntable = [atable[ix] for ix in range(len(atable))
              if atable[ix][1] is blank
              and not atable[ix][0].titl.split('.')[-1] in
              ['BeforeSuite', 'AfterSuite']
              ]

    ctable = [
        [headerCell(txt="Passed Test Count")],
        [headerCell(txt="Failed Test Count")],
        [headerCell(txt="Total")],
        [headerCell(txt="Not run Test Count")],
        [headerCell(txt="Failed Before/After Suite")],
        [headerCell(txt="Not run Before/After Suite")],
    ]
    for iex in range(len(executions)):
        # offset for header column
        ix = iex + 1
        fail_count = 0
        pass_count = 0
        notrun_count = 0
        beforeaftersuite_notrun_count = 0
        beforeaftersuite_fail_count = 0
        for iy in range(len(atable)):
            is_BA = atable[iy][0].titl.split(
                '.')[-1] in ['BeforeSuite', 'AfterSuite']
            if isinstance(atable[iy][ix], testRunCell):
                if atable[iy][ix].failed:
                    fail_count += 1
                    if is_BA:
                        beforeaftersuite_fail_count += 1
                else:
                    pass_count += 1
            else:
                if atable[iy][ix] is blank:
                    if not is_BA:
                        notrun_count += 1
                    else:
                        beforeaftersuite_notrun_count += 1
                else:
                    raise(Exception("unexpected cell"))
        ctable[0].append(plainCell(txt=f'{pass_count}'))
        ctable[1].append(plainCell(txt=f'{fail_count}'))
        ctable[2].append(plainCell(txt=f'{pass_count + fail_count}'))
        ctable[3].append(plainCell(txt=f'{notrun_count}'))
        ctable[4].append(plainCell(txt=f'{beforeaftersuite_fail_count}'))
        ctable[5].append(plainCell(txt=f'{beforeaftersuite_notrun_count}'))

    reportFile = os.path.join(
        _args.resultsdir,
        f'e2e-report-{_args.jiraKey}-{"-".join(testEnvironments)}.html'
    )
    print(f'Generating {reportFile}')
    with open(reportFile, 'w') as fp:
        startReport(fp, f'{_args.jiraKey}', testEnvironments)

        writeTable(fp, hdrRows, ftable, f'Failed tests ({len(ftable)})')

        writeTable(fp, hdrRows, gtable,
                   f'Test results graded on latest {score_sample_depth}'
                   'results, worst to best')

        writeTable(fp, hdrRows, ntable, f'Tests not run ({len(ntable)})')

        writeTable(fp, hdrRows, ttable, f'Tests ({len(ttable)})')

        writeTable(fp, hdrRows, atable,
                   'All tests (including BeforeSuite,AfterSuite,....)'
                   f' ({len(atable)})')

        # hack:
        #   - replace header Test\Day with Day
        #   - remove location column
        hdrRows[-1][0].txt = 'Day'
        hdrRows[-1] = hdrRows[-1][0:-1]
        writeTable(fp, hdrRows, ctable, 'Test count history')

        endReport(fp)


if __name__ == '__main__':
    from argparse import ArgumentParser
    import os

    parser = ArgumentParser(
        description='print test plan results.'
        ' Test results are printed in reverse chronological order'
        ' Query data can be cached using --outfile.'
        ' Cached query data can be used using --infile,'
        ' the xray query is skipped.'
    )

    _artefacts_dir = os.path.abspath(
        os.path.join(
            os.path.dirname(sys.argv[0]), '../artifacts'
        )
    )

    cachedir = os.path.join(_artefacts_dir, 'xray-report', 'cache')
    resultsdir = os.path.join(_artefacts_dir, 'xray-report', 'html')

    parser.add_argument('jiraKey', help='testplan JIRA key')
    parser.add_argument('--project', dest='project', default='MQ')
    parser.add_argument('--cachedir', dest='cachedir',
                        default=cachedir,
                        help='directory to cache json formatted results'
                        )
    parser.add_argument('--resultsdir', dest='resultsdir',
                        default=resultsdir,
                        help='directory for html formatted results'
                        )
    parser.add_argument('--collect', dest='collect',
                        action='store_true', default=False,
                        help='collect results from xray'
                        )
    parser.add_argument('--refresh', dest='refresh',
                        action='store_true', default=False,
                        help='refresh results from xray, ignore existing cache'
                        )
    parser.add_argument('--columns', dest='columns',
                        type=int, default=42,
                        help='number of results columns'
                        )
    parser.add_argument('--testenvs', dest='testenvs',
                        default=(
                            'osf_ubuntu,osv_20.04.3_lts,'
                            'osk_5.8.0-63-generic,'
                            'k8srel_1.21,'
                            'k8spatchrel_1.21.8,'
                            'plat_hcloud'
                        ),
                        help=', seperated list of test environments'
                        )

    _args = parser.parse_args()

    os.makedirs(_args.cachedir, exist_ok=True)
    os.makedirs(_args.resultsdir, exist_ok=True)

    testListFile = f'{_args.cachedir}/{_args.jiraKey}.tests.json'
    testExecsFile = f'{_args.cachedir}/{_args.jiraKey}.executions.json'

    try:
        if _args.collect:
            xrc = xraygraphql.XrayClient(project=_args.project)

            print(f'Collecting set of tests in testplan {_args.jiraKey}',
                  file=sys.stderr, end='')
            t0 = datetime.now()
            tests = xrc.GetTestsInTestPlan(jiraKey=_args.jiraKey)
            print(' (', datetime.now() - t0, ')', file=sys.stderr)
            jsonUpdate(testListFile, tests)

            print('Collecting set of test executions in testplan'
                  f' {_args.jiraKey}',
                  file=sys.stderr, end='')
            t1 = datetime.now()
            executions = xrc.GetTestExecutionsInTestPlan(jiraKey=_args.jiraKey)
            print(' (', datetime.now() - t1, ')', file=sys.stderr)
            jsonUpdate(testExecsFile, executions)

            print(f'Collecting set of test runs in testplan {_args.jiraKey}: ',
                  file=sys.stderr, end='')
            t2 = datetime.now()
            for jiraKey, testExec in executions.items():
                execFile = f'{_args.cachedir}/exec.{jiraKey}.json'
                if _args.refresh or not jsonExists(execFile):
                    t3 = datetime.now()
                    print(f'{jiraKey} ', file=sys.stderr, end='')
                    sys.stderr.flush()
                    execData = xrc.GetTestExecutionRuns(
                        issueId=testExec['issueId'])
                    print(' (', datetime.now() - t3, ')', file=sys.stderr)
                    jsonSave(execFile, execData)
                    # sleep to workaround
                    #       TRACE: request failed 429, Too Many Requests
                    sleep(0.5)
            print(' (', datetime.now() - t2, ')', file=sys.stderr)

        main()
    except Exception:
        traceback.print_exc()
