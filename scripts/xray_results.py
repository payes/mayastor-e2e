#!/usr/bin/env python3
import sys
import json
from pprint import pprint
from argparse import ArgumentParser
import xraygraphql
import analyse
import os
from datetime import datetime, timedelta
import re
from io import StringIO
from time import sleep


def handle_testplan(testRuns, opts):
    ''' pretty printing is hard :-(
    '''

    def getCutoffTs(date):
        datestr = date
        datestr = datestr.replace('/', '-')
        midniteUtc = datetime.utcnow().replace(
            hour=0, minute=0, second=0, microsecond=0)

        if datestr == '0d':
            return midniteUtc.timestamp() * 1000

        # fixme all datetime objects should be utc
        dt = None
        try:
            dt = datetime.strptime(datestr, "%dd")
            td = timedelta(days=dt.day)
        except ValueError as ve:
            _ = ve
            # print(ve, file=sys.stderr)
            pass

        try:
            dt = datetime.strptime(datestr, "%dw")
            td = timedelta(days=dt.day*7)
        except ValueError as ve:
            _ = ve
            # print(ve, file=sys.stderr)
            pass

        if dt is not None:
            dt = midniteUtc - td
            return dt.timestamp() * 1000

        for fmt in [
            '%Y-%m-%d',
            '%d-%m-%y',
            '%d-%m-%Y',
        ]:
            try:
                return datetime.strptime(datestr, fmt).timestamp()*1000
            except ValueError as ve:
                _ = ve
                # print(ve, file=sys.stderr)
                pass

        return -1

    def printResults(strm, tRuns, tid):
        printedHeader = False
        ix = 1
        tmp = {}
        for run in tRuns:
            tmp[run['finishedOn']] = run
        for k in sorted(tmp.keys(), reverse=True):
            run = tmp[k]
            synopsis = (
                f'{ix}){run["finishedOnHR"]},'
                f' {run["status"]["name"]},'
                f' id={run["id"]}'
            )
            failed = run['status']['name'] == 'FAILED'

            if opts.onlyFails and not failed:
                continue

            if not printedHeader:
                print(f'{tid} {run["e2e_name"]}:', file=strm)
                print(f'{run["unstructured"]}', file=strm)
                printedHeader = True

            print(synopsis, file=strm)

            if opts.dump:
                pprint(run)
                print('', file=strm)
            elif opts.verbose and failed:
                for res in run['results']:
                    print(res['log'], file=strm)
                print('', file=strm)
            ix += 1
        del tmp

    def newFails(strm, tRuns, tid):
        tmp = {}
        for run in tRuns:
            tmp[run['finishedOn']] = run
        kiis = sorted(tmp.keys(), reverse=True)
        del run
        if len(kiis) == 0:
            return
        run = tmp[kiis[0]]

        if run['status']['name'] != 'FAILED':
            return
        synopsis = (
            f' {run["finishedOnHR"]},'
            f' {run["status"]["name"]},'
            f' id={run["id"]}'
        )
        try:
            run2 = tmp[kiis[1]]
            if run2['status']['name'] == 'FAILED':
                return
        except IndexError:
            pass

        print(f'{tid} {run["e2e_name"]}:', file=strm)
        print(f'{run["unstructured"]}', file=strm)
        print(synopsis, file=strm)

        return run
        del tmp

    def summariseResults(strm, tRuns, tid):
        summary = []
        # pah! printing is hard, return if nothing to do.
        if 0 == len(tRuns):
            return summary

        ix = 0
        count = 0
        run = tRuns[-1]
        prev_status = run['status']['name']

        failed = False
        for run in tRuns[::-1]:
            status = run['status']['name']
            if run['status']['name'] == 'FAILED':
                failed = True
            if status == prev_status:
                count += 1
            else:
                summary.append({prev_status: count})
                prev_status = status
                count = 1
            ix += 1

        if count != 0:
            summary.append({status: count})

        if not opts.onlyFails or failed:
            print(
                f'{tid} ({run["e2e_name"]})\n{run["unstructured"]}:\n\t',
                end='', file=strm)
            for entry in summary:
                for k, v in entry.items():
                    print(f'{k}:{v} ', end='', file=strm)
            print('', file=strm)
            return summary
        return []

    def tableResults(results, tests, index):
        rindex = {}
        for k, v in index.items():
            for t in v:
                rindex[t] = k

        test_list = sorted(
            tests.keys(), key=lambda test: int(test.split('-')[-1]))
        print(test_list)
        e2enames = [rindex.get(t) for t in test_list]
        print(sorted(set(e2enames)))

    results = testRuns['results']
    filter = []
    if opts.testJiraKeys:
        testJiraKeys = re.split(',| ', opts.testJiraKeys)
        filter.extend(testJiraKeys)

    def fix_index():
        ''' Heuristic fixes, to group results correctly
        '''
        reindex_map = {
            'CSI E2E Suite': 'csi',
            'Primitive MSP stress test': 'primitive_msp_stress',
            'csi,resource_check': 'csi'
        }

        index = {}
        for k, v in testRuns['index'].items():
            ak = k.split('.')[0]
            ak = reindex_map.get(ak, ak)
            try:
                index[ak].extend(v)
            except KeyError:
                index[ak] = v

        return index

    index = fix_index()

    if opts.e2enames:
        for x in re.split(',| ', opts.e2enames):
            filter.extend(index[x])

    if 0 != len(filter):
        results = {k: v for k, v in results.items() if k in filter}

    if opts.date:
        ts = getCutoffTs(opts.date)
        if ts < 0:
            raise(Exception(f'unsupported time format {opts.date}'))
        results = {
            jk: [run for run in runz if int(run['finishedOn']) >= ts]
            for jk, runz in results.items()
        }
        # remove entries with no results
        results = {jk: runz for jk, runz in results.items() if len(runz) > 0}

    if opts.tabular:
        tableResults(results, testRuns['tests'], index)
        return

    if opts.group:
        for e2ename, tl in index.items():
            tmp = [x for x in tl if x in results]
            if len(tmp) == 0:
                continue
            graded = []
            strm = StringIO('')
            for jk in sorted(tmp, key=lambda test: int(test.split('-')[-1])):
                if opts.summarise:
                    summariseResults(strm, results[jk], jk)
                if opts.print:
                    printResults(strm, results[jk], jk)
                if opts.newfails:
                    newFails(strm, results[jk], jk)
                if opts.grade:
                    graded.append(summariseResults(strm,
                                                   results[jk], jk))
            if opts.grade:
                g0 = []
                for grade in graded:
                    try:
                        g0.append(grade[0])
                    except IndexError as ie:
                        print(e2ename, grade, graded)
                        raise(ie)
                sg0 = sorted(g0, key=lambda x: x.get(
                    'PASSED', -1 * x.get('FAILED', 1)))
                if opts.gradepass == 0 or \
                        sg0[0].get('PASSED', 0) >= opts.gradepass:
                    print(f'{e2ename}:')
                    print(sg0[0], ' :', g0, '\n')
            else:
                content = strm.getvalue()
                if len(content) != 0:
                    print(f'{e2ename}:')
                    print(content)
            strm.truncate(0)
    else:
        for jk in sorted(results.keys()):
            if opts.summarise:
                summariseResults(strm, results[jk], jk)
            if opts.print:
                printResults(strm, results[jk], jk)
            if opts.newfails:
                newFails(strm, results[jk], jk)
            print(strm.getvalue())
            strm.truncate(0)


if __name__ == '__main__':
    parser = ArgumentParser(
        description='print test plan results.'
        ' Test results are printed in reverse chronological order'
        ' Query data can be cached using --outfile.'
        ' Cached query data can be used using --infile,'
        ' the xray query is skipped.'
    )

    parser.add_argument('jiraKey', help='testplan JIRA key')
    parser.add_argument('--project', dest='project', default='MQ')
    parser.add_argument('-o', '--outfile', dest='outfile',
                        default=None,
                        help='file to cache json formatted results'
                        )
    parser.add_argument('-i', '--infile', dest='infile',
                        default=None,
                        help='file to read previously cached'
                        ' json formatted results'
                        ' if this option is specified,'
                        ' the xray graphql query is skipped'
                        ' results cached in the file are used'
                        )
    parser.add_argument('-t', '--testJiraKeys',
                        dest='testJiraKeys', default=None,
                        help='filter: comma separated list of test JIRA keys'
                        'process results for these tests only'
                        )
    parser.add_argument('-e', '--e2enames',
                        dest='e2enames', default=None,
                        help='filter: comma separated list of e2e test names'
                        'process results for these testsuites only'
                        )
    parser.add_argument('-f', '--fails', dest='onlyFails',
                        action='store_true', default=False,
                        help='filter: print failed tests only'
                        )
    parser.add_argument('-D', '--date', dest='date',
                        default=None,
                        help='filter: cut off date %%Y-%%m-%%d,'
                             '%%dd (days back), %%ww (weeks back)'
                        )
    parser.add_argument('-p', '--print', dest='print',
                        action='store_true', default=False,
                        help='print results'
                        )
    parser.add_argument('-v', '--verbose', dest='verbose',
                        action='store_true', default=False,
                        help='print execution log of test'
                        )
    parser.add_argument('-d', '--dump', dest='dump',
                        action='store_true', default=False,
                        help='dump full record of test'
                        )
    parser.add_argument('-s', '--summarise', dest='summarise',
                        action='store_true', default=False,
                        help='summarise results'
                        )
    parser.add_argument('--tabular', dest='tabular',
                        action='store_true', default=False,
                        help='tabular results'
                        )
    parser.add_argument('--ungroup', dest='group',
                        action='store_false', default=True,
                        help='do not group results on e2ename'
                        )
    parser.add_argument('--grade', dest='grade',
                        action='store_true', default=False,
                        help='grade results on e2ename'
                        )
    parser.add_argument('--gradepass', dest='gradepass',
                        type=int, default=0,
                        help='grade pass value (default = 0)'
                        )
    parser.add_argument('--newfails', dest='newfails',
                        action='store_true', default=False,
                        help='list new failures'
                        )

    _args = parser.parse_args()

    if _args.grade and _args.onlyFails:
        print('--grade incompatible with --fails', file=sys.stderr)
        exit(1)

    if _args.infile is not None:
        if _args.outfile is not None:
            print('--infile trumps --outfile, zapping --outfile',
                  file=sys.stderr)
            _args.outfile = None
        with open(_args.infile, mode='r') as fp:
            data = json.load(fp)[_args.jiraKey]
    else:
        xrc = xraygraphql.XrayClient(project=_args.project)

        print(f'Collecting set of tests in testplan {_args.jiraKey}',
              file=sys.stderr, end='')
        t0 = datetime.now()
        tests = xrc.GetTestsInTestPlan(jiraKey=_args.jiraKey)
        print(' (', datetime.now() - t0, ')', file=sys.stderr)

        print(f'Collecting set of test executions in testplan {_args.jiraKey}',
              file=sys.stderr, end='')
        t1 = datetime.now()
        executions = xrc.GetTestExecutionsInTestPlan(jiraKey=_args.jiraKey)
        testExecs = xrc.GetTestExecutionsInTestPlan(jiraKey=_args.jiraKey)
        print(' (', datetime.now() - t1, ')', file=sys.stderr)

        print(f'Collecting set of test runs in testplan {_args.jiraKey}: ',
              file=sys.stderr, end='')
        t2 = datetime.now()
        testRuns = []
        for jiraKey, testExec in testExecs.items():
            # print(f'Trace:{jiraKey} : {testExec['issueId']}',
            # file=sys.stderr)
            print(f'{jiraKey} ', file=sys.stderr, end='')
            sys.stderr.flush()
            testRuns.extend(xrc.GetTestExecutionRuns(
                issueId=testExec['issueId']))
            # sleep to workaround TRACE: request failed 429, Too Many Requests
            sleep(0.5)
        print(' (', datetime.now() - t2, ')', file=sys.stderr)

        # annotate the results use 'unstructured' => e2e_name, e2e_file
        defs2src = analyse.scrapeSources(
            os.path.realpath(sys.path[0] + '/../src'))
        defs2src = {
            k: {
                'e2e_name': v.split('/')[-2],
                'e2e_file': v.split('/')[-1]
            } for k, v in defs2src.items()
        }
        for tr in testRuns:
            tr.update(defs2src.get(tr.get('unstructured'),
                      {'e2e_name': tr.get('unstructured'), 'e2e_file': '?'}))

        # Convert the array of results into a map keyed on jira key
        # Create and index from e2e test name to Jira key
        results = {}
        index = {}
        for tr in testRuns:
            k = tr['test']['jira']['key']
            try:
                index[tr['e2e_name']].append(k)
            except KeyError:
                index[tr['e2e_name']] = [k]
            try:
                results[k].append(tr)
            except KeyError:
                results[k] = [tr]
        for k, v in index.items():
            index[k] = sorted(set(v))
        data = {
            'results': results,
            'index': index,
            'tests': tests,
            'executions': executions,
        }

    if _args.outfile is not None:
        with open(_args.outfile, 'w') as fp:
            json.dump({_args.jiraKey: data}, fp, sort_keys=True, indent=4)

    handle_testplan(data, _args)
