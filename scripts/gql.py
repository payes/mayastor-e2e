#!/usr/bin/env python3
import sys
import json
from pprint import pprint, pformat
from argparse import ArgumentParser
import xraygraphql
import analyse
import os


def dict_print(d, indent=""):
    kiis = sorted(d.keys())
    for k in kiis:
        v = d[k]
        if isinstance(v, dict):
            print(f'{indent}{k}: ')
            dict_print(v, indent + "    ")
        else:
            print(f'{indent}{k}: {v}')


def action_list_tests():
    parser = ArgumentParser()
    parser.add_argument('xraytype')
    parser.add_argument('-p', '--project', dest='project', default='MQ')
    parser.add_argument('-status', '--status', dest='status', default=None,
                        help='comma separated list of status')
    parser.add_argument('--nots', dest='emptyTestSets',
                        action='store_true', default=False,
                        help='filter: print tests not in any test sets'
                        )
    parser.add_argument('--notp', dest='emptyTestPlans',
                        action='store_true', default=False,
                        help='filter: print tests not in any test plans'
                        )
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

    _args = parser.parse_args()

    if _args.infile is not None:
        if _args.outfile is not None:
            print('--infile trumps --outfile, zapping --outfile',
                  file=sys.stderr)
            _args.outfile = None
        with open(_args.infile, mode='r') as fp:
            tests = json.load(fp)
    else:
        xrc = xraygraphql.XrayClient(project=_args.project)
        tests = xrc.ListTests()
        if _args.outfile is not None:
            with open(_args.outfile, 'w') as fp:
                json.dump(tests, fp, sort_keys=True, indent=4)

    include_status = []
    exclude_status = []
    if _args.status is not None:
        include_status = _args.status.split(',')
        exclude_status = [x[1:] for x in include_status if x.startswith("!")]
        include_status = [x for x in include_status if not x.startswith("!")]
    for mq in sorted(tests.keys(), key=lambda x: int(x.split('-')[-1])):
        entry = tests[mq]
        entry['jira'] = {k: v for k, v in entry['jira'].items() if k not in [
            'key']}
        entry['jira']['status'] = {
            k: v for k, v in entry['jira']['status'].items()
            if k in ['name', 'description']
        }

        # filters
        if len(include_status) != 0 and entry['jira']['status']['name'] \
                not in include_status:
            continue
        if len(exclude_status) != 0 and entry['jira']['status']['name'] \
                in exclude_status:
            continue
        if _args.emptyTestSets and len(entry['testSets']['results']) != 0:
            continue
        if _args.emptyTestPlans and len(entry['testPlans']['results']) != 0:
            continue

        print(f'{mq}: ')
        # prettify for display
        entry['testSets'] = [e['jira']['key']
                             for e in entry['testSets']['results']]
        entry['testPlans'] = [e['jira']['key']
                              for e in entry['testPlans']['results']]
        dict_print(entry, "   ")


def action_list_testplans():
    parser = ArgumentParser()
    parser.add_argument('xraytype')
    parser.add_argument('action')
    parser.add_argument('-p', '--project', dest='project', default='MQ')
    _args = parser.parse_args()
    xrc = xraygraphql.XrayClient(project=_args.project)
    testPlans = xrc.ListTestPlans()
    for mq, entry in testPlans.items():
        print(f'{mq}: {entry["jira"]["summary"]}\n'
              f'description:\n{entry["jira"]["description"]}'
              '\n'
              )


def action_list_testplan_tests():
    # parse args again for getkubeconfig specifics
    parser = ArgumentParser()
    parser.add_argument('xraytype')
    parser.add_argument('action')
    parser.add_argument('jiraKey', help='testplan JIRA key')
    parser.add_argument('-p', '--project', dest='project', default='MQ')
    _args = parser.parse_args()

    xrc = xraygraphql.XrayClient(project=_args.project)

    tests = xrc.GetTestsInTestPlan(jiraKey=_args.jiraKey)
    for mq in sorted(tests.keys(), key=lambda x: int(x.split('-')[-1])):
        entry = tests[mq]
        entry['jira'] = {k: v for k,
                         v in entry['jira'].items() if k not in 'key'}
        print(f'{mq}: {pformat(entry)}')


def action_list_testplan_executions():
    # parse args again for getkubeconfig specifics
    parser = ArgumentParser()
    parser.add_argument('xraytype')
    parser.add_argument('action')
    parser.add_argument('jiraKey', help='testplan JIRA key')
    parser.add_argument('-p', '--project', dest='project', default='MQ')
    _args = parser.parse_args()

    xrc = xraygraphql.XrayClient(project=_args.project)

    testExecs = xrc.GetTestExecutionsInTestPlan(jiraKey=_args.jiraKey)
    for mq in sorted(testExecs.keys(), key=lambda x: int(x.split('-')[-1])):
        entry = testExecs[mq]
        print(f'{mq}: {pformat(entry)}')


def action_testplan_grade():
    parser = ArgumentParser()
    parser.add_argument('xraytype')
    parser.add_argument('action')
    parser.add_argument('jiraKey', help='JIRA key of testplan')
    parser.add_argument('-p', '--project', dest='project', default='MQ')
    parser.add_argument('-o', dest='outdir', default=None,
                        help='directory to output sets of data in json format')
    _args = parser.parse_args()

    xrc = xraygraphql.XrayClient(project=_args.project)

    def _opath(basename):
        return os.path.join(_args.outdir, basename)

    testRuns = []
    exes = xrc.GetTestExecutionsInTestPlan(jiraKey=_args.jiraKey)
    for k, exe in exes.items():
        print('.', file=sys.stderr, end='')
        sys.stderr.flush()
        testRuns.extend(xrc.GetTestExecutionRuns(issueId=exe['issueId']))

    print('', file=sys.stderr)
    if _args.outdir is not None:
        with open(_opath(f'{_args.jiraKey}.testruns.json'), 'w') as fp:
            json.dump(testRuns, fp, sort_keys=True, indent=4)

    mqRuns, defRuns, mq2def = analyse.testRunComprehensions(testRuns)

    # get mapping from definitions to source
    defs2src = analyse.scrapeSources(os.path.realpath(sys.path[0] + '/../src'))

    if _args.outdir is not None:
        with open(_opath(f'{_args.jiraKey}.defs2src.json'), 'w') as fp:
            json.dump(defs2src, fp, sort_keys=True, indent=4)

        srcRuns = {defs2src[k].split('/')[-2]: defRuns[k]
                   for k in defRuns if k in defs2src}
        comp = {
            'ByDefinition': defRuns,
            'ByJiraKey': mqRuns,
            'BySrc': srcRuns,
            'JiraKey2Definition': mq2def
        }

        with open(_opath(f'{_args.jiraKey}.comp.json'), 'w') as fp:
            json.dump(comp, fp, sort_keys=True, indent=4)

        with open(_opath(f'{_args.jiraKey}.definitions.json'), 'w') as fp:
            json.dump(defRuns, fp, sort_keys=True, indent=4)
#    analyse.printDefRuns(defRuns)

    gradedDefs = analyse.gradeDefRuns(defRuns)
    if _args.outdir is not None:
        with open(_opath(f'{_args.jiraKey}.definition.grades.json'),
                  'w') as fp:
            json.dump(gradedDefs, fp, sort_keys=True, indent=4)

    srcGrades = analyse.gradedSrcRuns(gradedDefs, defs2src)
    if _args.outdir is not None:

        with open(_opath(f'{_args.jiraKey}.src.grades.json'), 'w') as fp:
            json.dump(srcGrades, fp, sort_keys=True, indent=4)
    analyse.printGradedSrcRuns(srcGrades)


def action_list_testexecutions():
    parser = ArgumentParser()
    parser.add_argument('xraytype')
    parser.add_argument('action')
    parser.add_argument('-p', '--project', dest='project', default='MQ')
    _args = parser.parse_args()
    xrc = xraygraphql.XrayClient(project=_args.project)
    testRuns = xrc.ListTestExecutions()
    pprint(testRuns)


def action_list_testexecution_runs():
    parser = ArgumentParser()
    parser.add_argument('xraytype')
    parser.add_argument('action')
    parser.add_argument('jiraKey', help='JIRA key of testexecution')
    parser.add_argument('-p', '--project', dest='project', default='MQ')
    _args = parser.parse_args()

    xrc = xraygraphql.XrayClient(project=_args.project)

    testRuns = xrc.GetTestExecutionRuns(jiraKey=_args.jiraKey)
    pprint(testRuns)


def action_list_testruns():
    parser = ArgumentParser()
    parser.add_argument('action')
    parser.add_argument('jiraKey', help='JIRA key of test')
    parser.add_argument('-p', '--project', dest='project', default='MQ')
    parser.add_argument('--failed', dest='failed_only',
                        action='store_true', default=False,
                        help='only print failed test runs')
    parser.add_argument('--succinct', dest='succinct',
                        action='store_true', default=False,
                        help='succinct print out')
    _args = parser.parse_args()

    xrc = xraygraphql.XrayClient(project=_args.project)

    testRuns = xrc.GetTestRuns(jiraKey=_args.jiraKey)
    if _args.failed_only:
        for tr in testRuns:
            if tr['status']['name'] == 'FAILED':
                if _args.succinct:
                    print('{')
                    print(tr['datetime'], tr['id'])
                    for res in tr['results']:
                        print(res['log'])
                    print('}\n')
                else:
                    pprint(tr)
    else:
        pprint(testRuns)


if __name__ == '__main__':
    # Arguments handling is bound to offend someone :-(
    # For now use a single script to service different actions
    # Arguments parsing is 2 phase,
    # 1 - figure out what the base action is call helper function to
    #     parse arguments for the base action
    # 2 - helper function parses arguments again as appropriate for the
    #     action
    # Pros: advantage simpler argument parsing (for an action)
    # Cons: help doesn't work
    parser = ArgumentParser(add_help=False)
    parser.add_argument('xraytype', choices=[
        'testplan',
        'testsets',
        'testexecutions',
        'testruns',
        'tests',
    ])

    # parse args only for xraytype
    _args, _ = parser.parse_known_args()

    if _args.xraytype == 'tests':
        action_list_tests()
    if _args.xraytype == 'testplan':
        parser.add_argument('action', choices=[
            'list',
            'tests',
            'executions',
            'grade',
        ])
        _args, _ = parser.parse_known_args()
        action = _args.action
        if action == 'list':
            action_list_testplans()
        elif action == 'tests':
            action_list_testplan_tests()
        elif action == 'executions':
            action_list_testplan_executions()
        elif action == 'grade':
            action_testplan_grade()
    elif _args.xraytype == 'testexecutions':
        parser.add_argument('action', choices=['list', 'runs', ])
        _args, _ = parser.parse_known_args()
        action = _args.action
        if action == 'list':
            action_list_testexecutions()
        elif action == 'runs':
            action_list_testexecution_runs()
    elif _args.xraytype in ['testruns']:
        action_list_testruns()
