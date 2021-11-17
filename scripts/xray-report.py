#!/usr/bin/env python3
from datetime import datetime
import html
import json
import sys
from time import sleep
import xraygraphql

htmlStart = '''
<!doctype html>

<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">

  <title>Mayastor E2E Results</title>
  <h1>Mayastor E2E Results</h1>
  <meta name="description" content="Mayastor E2E test results.">
  <meta name="author" content="xray-report.py">

  <link rel="stylesheet" href="css/styles.css?v=1.0">

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
  a {text-decoration: none;}
  </style>

</head>

<body style="background-color:#d4d4d4;">
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
    def __init__(self, txt='', link=None, hover=None, span=0,
                 isHeader=False, ts=0, passFail=True):
        self.txt = txt
        self.link = link
        self.hover = hover
        # span is column span
        self.span = span
        self.isHeader = isHeader
        self.passFail = True

        # ts is not used for display
        self.ts = ts

    def __str__(self):
        return f'{self.txt} {self.ts}'

    def _td(self):
        f = []
        f.append('<td')

        if self.hover:
            hover = html.escape(self.hover)
            f.append(f' title="{hover}"')

        txt = self.txt
        if self.passFail:
            f.append(' style="text-align: center; vertical-align: middle;"')
            if self.txt == 'F':
                f.append(' bgcolor="#FFD1DC")')
                txt = '&#x2713;'
                # f.append(' bgcolor="#f8d2d2")')
            elif self.txt == 'P':
                f.append(' bgcolor="#93E9BE")')
                # f.append(' bgcolor="#e5ffe5")')
                txt = '&#x2717;'
            else:
                f.append(' bgcolor="#c8c8c8")')
                txt = '&#8205;'
            # Invisible, but clickable
            txt = '<span style="opacity:0;">00</span>'

        f.append('>')

        if self.link:
            f.append(f'<a href="{self.link}" target="_blank">{txt}</a>')
        else:
            f.append(f'{txt}')

        f.append('</td>')

        return ''.join(f)

    def _th(self):
        f = []
        f.append('<th')

        if self.hover:
            f.append(f' title="{self.hover}"')

        if self.span:
            f.append(f' colspan="{self.span}"')

        f.append('>')

        if self.link:
            f.append(f'<a href="{self.link}">{self.txt}</a>')
        else:
            f.append(f'{self.txt}')

        f.append('</th>')

        return ''.join(f)

    def render(self):
        if self.isHeader:
            return self._th()
        return self._td()


def issueLink(issueKey):
    return f'https://mayadata.atlassian.net/browse/{issueKey}'


def runLink(execKey, issueKey):
    # https://mayadata.atlassian.net/plugins/servlet/ac/com.xpandit.plugins.xray/execution-page?ac.testExecIssueKey=MQ-2532&ac.testIssueKey=MQ-2425
    return ('https://mayadata.atlassian.net/plugins/servlet/ac/'
            'com.xpandit.plugins.xray/execution-page?'
            f'ac.testExecIssueKey={execKey}'
            f'&ac.testIssueKey={issueKey}')


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

    if not os.path.isdir(_artefacts_dir):
        cachedir = '/tmp/xray-report/cache'
        resultsdir = '/tmp/xray-report/html'
    else:
        cachedir = os.path.join(_artefacts_dir,'xray-report','cache')
        resultsdir = os.path.join(_artefacts_dir,'xray-report','html')

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

    _args = parser.parse_args()

    os.makedirs(_args.cachedir, exist_ok=True)
    os.makedirs(_args.resultsdir, exist_ok=True)

    testListFile = f'{_args.cachedir}/{_args.jiraKey}.tests.json'
    testExecsFile = f'{_args.cachedir}/{_args.jiraKey}.executions.json'

    if _args.collect:
        xrc = xraygraphql.XrayClient(project=_args.project)

        print(f'Collecting set of tests in testplan {_args.jiraKey}',
              file=sys.stderr, end='')
        t0 = datetime.now()
        tests = xrc.GetTestsInTestPlan(jiraKey=_args.jiraKey)
        print(' (', datetime.now() - t0, ')', file=sys.stderr)
        jsonUpdate(testListFile, tests)

        print(f'Collecting set of test executions in testplan {_args.jiraKey}',
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

    # Objective create a 2 dimensional array from which rendering as an
    # html table is a prettifying but dumb op
    # This means filling missing test runs with placeholders
    # which render as empty table cells

    tests = jsonLoad(testListFile)
    testKeys = sorted(
        tests.keys(),
        key=lambda test: int(test.split('-')[-1])
    )

    executions = [v for _, v in jsonLoad(testExecsFile).items()]

    blank = tableCell()
    blanks = [blank for _ in executions]

    # reverse for reverse chronological order
    # first attempt used 'lastModified', but that gives spurious
    # ordering, and there is no equivalent og 'createdTime'
    # so we order on MQ-number
    executions = sorted(executions,
                        key=lambda d: int(d['jira']['key'].split('-')[-1]),
                        reverse=True)

    htable = [[
        tableCell(
            hover=d['jira']['summary'],
            link=issueLink(d['jira']['key']),
            isHeader=True
        )
        for d in executions
    ]]

    # column 0
    htable[0].insert(0, tableCell(isHeader=True))
    row = 1
    for testKey in testKeys:
        test = tests[testKey]
        htable.append(
            [tableCell(
                txt=testKey,
                hover=test['unstructured'],
                link=issueLink(testKey),
                isHeader=True)
             ])
        htable[row].extend(blanks)
        row += 1

    col = 1
    for exe in executions:
        execDict = {
            d['test']['jira']['key']: d for d in jsonLoad(
                f'{_args.cachedir}/exec.{exe["jira"]["key"]}.json')
        }
        row = 1
        ts = 0
        for testKey in testKeys:
            if testKey in execDict:
                test = execDict[testKey]
                htable[row][col] = tableCell(
                    txt=test['status']['name'][0],
                    hover=test['results'][0]['log'],
                    link=runLink(exe["jira"]["key"], testKey),
                    passFail=True
                )

                finished = int(test['finishedOn'])
                if ts == 0 or finished < ts:
                    ts = finished
            row += 1
        htable[0][col].ts = int(ts/1000)
        col += 1

    dt = datetime.fromtimestamp(float(htable[0][1].ts))
    year_span = [{'txt': dt.strftime('%Y'), 'span': 0}]
    month_span = [{'txt': dt.strftime('%b'), 'span': 0}]
    days = []

    col = 1
    for x in htable[0][1:]:
        dt = datetime.fromtimestamp(float(x.ts))
        dt.replace(second=0, microsecond=0)
        year = dt.strftime('%Y')
        if year == year_span[-1]['txt']:
            year_span[-1]['span'] += 1
        else:
            year_span.append({'txt': year, 'span': 1})

        mon = dt.strftime('%b')
        if mon == month_span[-1]['txt']:
            month_span[-1]['span'] += 1
        else:
            month_span.append({'txt': mon, 'span': 1})
        days.append(dt.strftime('%d'))
        htable[0][col].txt = dt.strftime('%d')
        col += 1

    reportFile = os.path.join(_args.resultsdir, 'e2e-report.html')
    print(f'Generating {reportFile}')
    with open(reportFile, 'w') as fp:
        print(htmlStart, file=fp)
        tplink = issueLink(_args.jiraKey)
        print(f'<h2>Testplan <a href="{tplink}">{_args.jiraKey}</a></h2>', file=fp)
        print('<table>', file=fp)
        print('<tr>', file=fp)

        # headers
        print('<th>Year</th>', file=fp)
        for cell in year_span:
            print(f'<th colspan="{cell["span"]}"]>{cell["txt"]}</th>', file=fp)
        print('</tr>', file=fp)

        print('<tr>', file=fp)
        print('<th>Month</th>', file=fp)
        for cell in month_span:
            print(f'<th colspan="{cell["span"]}"]>{cell["txt"]}</th>', file=fp)
        print('</tr>', file=fp)

#    print('<tr>', file=fp)
#    for col in htable[0]:
#        print(col.th(), file=fp)
#    print('</tr>', file=fp)

#    for row in htable[1:]:
        for row in htable:
            print('<tr>', file=fp)
            print(row[0].render(), file=fp)
            for col in row[1:]:
                print(col.render(), file=fp)
            print('</tr>', file=fp)

        print('</table>', file=fp)
        print(htmlEnd, file=fp)
