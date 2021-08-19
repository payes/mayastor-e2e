#!/usr/bin/env python3
import sys
import json
from argparse import ArgumentParser
import analyse
import os
from datetime import datetime


if __name__ == '__main__':
    parser = ArgumentParser(
        description='test check.'
    )
    parser.add_argument('--tests', dest='testsfile',
                        default=None,
                        help='file to read previously cached'
                        ' tests json formatted results'
                        ' if this option is specified,'
                        ' the xray graphql query is skipped'
                        ' results cached in the file are used'
                        )
    parser.add_argument('--testsets', dest='testsetsfile',
                        default=None,
                        help='file to read previously cached'
                        ' testset json formatted results'
                        ' if this option is specified,'
                        ' the xray graphql query is skipped'
                        ' results cached in the file are used'
                        )
    _args = parser.parse_args()

    defs2src = analyse.scrapeSources(
        os.path.realpath(sys.path[0] + '/../src'))
    defs2src = {
        k: {
            'e2e_name': v.split('/')[-2],
            'e2e_file': v.split('/')[-1],
            'classname': k.split('.')[0]
        } for k, v in defs2src.items()
    }

    with open(_args.testsfile, mode='r') as fp:
        tests = json.load(fp)

    with open(_args.testsetsfile, mode='r') as fp:
        testsets = json.load(fp)

    cn2ts = {}
    for mq, v in testsets.items():
        for src in v['sources']:
            cn2ts[src['classname']] = mq

    nodefs = {k: v for k, v in tests.items() if v['unstructured'] is None}
    print('=============================================================')
    print('Tests with no definitions')
    print('=============================================================')
    for test, entry in sorted(
        nodefs.items(), key=lambda x: int(x[0].split('-')[-1])
    ):
        jira = entry['jira']
        print(f'{test}:'
              f'\n\tstatus: {jira["status"]["name"]}'
              f'\n\tsummary: {jira["summary"]}'
              )
    print('')

    print('=============================================================')
    print('Tests (In Use, To Do) not belonging to any testPlan')
    print('=============================================================')
    notp = {
        k: v for k, v in tests.items()
        if len(v['testPlans']['results']) == 0
        and k not in nodefs
        and v['jira']['status']['name'] in ['In Use', 'To Do']
    }
    for test, entry in sorted(
        notp.items(), key=lambda x: int(x[0].split('-')[-1])
    ):
        jira = entry['jira']
        classname = entry['unstructured'].split('.')[0]
        print(f'{test}:'
              f'\n\tstatus: {jira["status"]["name"]}'
              f'\n\tdefinition: {entry["unstructured"]}'
              )

    print('=============================================================')
    print('Tests (In Use, To Do) not belonging to any testset')
    print('=============================================================')
    nots = {
        k: v for k, v in tests.items()
        if len(v['testSets']['results']) == 0
        and k not in nodefs
        and v['jira']['status']['name'] in ['In Use', 'To Do']
    }

    for test, entry in sorted(
        nots.items(), key=lambda x: int(x[0].split('-')[-1])
    ):
        jira = entry['jira']
        classname = entry['unstructured'].split('.')[0]
        candidate_ts = cn2ts.get(classname)
        if candidate_ts is not None:
            print(f'{test}:'
                  f'\n\tstatus: {jira["status"]["name"]}'
                  )
            print(f'\tclassname: {classname}')
            print(f'\tshould be in testset: {cn2ts.get(classname)}')

    print('------------------------')

    for test, entry in sorted(
        nots.items(), key=lambda x: int(x[0].split('-')[-1])
    ):
        jira = entry['jira']
        classname = entry['unstructured'].split('.')[0]
        candidate_ts = cn2ts.get(classname)
        if candidate_ts is None:
            print(f'{test}:'
                  f'\n\tstatus: {jira["status"]["name"]}'
                  )
            print(f'\tclassname: {classname}')
            print('\tfailed to find testset using exist test <-> testSets:'
                  f'\n\tdef:{entry["unstructured"]}'
                  )

    print('=============================================================')
    print('Tests with status TODO and have run(s) ')
    print('=============================================================')
    todo = {k: v for k, v in tests.items()
            if v['jira']['status']['name'] == 'To Do'}
    for test, entry in sorted(
        todo.items(), key=lambda x: int(x[0].split('-')[-1])
    ):
        jira = entry['jira']
        runs = [datetime.fromtimestamp(float(x['finishedOn'])/1000).ctime()
                for x in entry['testRuns']['results']]
        if len(runs) == 0:
            continue
        print(f'{test}:'
              f'\n\tstatus: {jira["status"]["name"]}'
              )
        try:
            print(f'\tlatest run: {runs[0]}')
        except IndexError:
            print('\tlatest run:')
        print(f'\ttestPlan: {entry["testPlans"]["results"]}')
        print(f'\tdefinition: {entry["unstructured"]}')
