#!/usr/bin/env python3
import sys
import json
from argparse import ArgumentParser
import xraygraphql
import analyse
import os


if __name__ == '__main__':
    parser = ArgumentParser()
    parser.add_argument('-p', '--project', dest='project', default='MQ')
    parser.add_argument('-o', dest='outfile', default=None)
    parser.add_argument('-i', dest='infile', default=None)
    _args = parser.parse_args()

    defs2src = analyse.scrapeSources(
        os.path.realpath(sys.path[0] + '/../src'))
    defs2src = {
        k: {
            'location': v,
            'classname': k.split('.')[0],
        } for k, v in defs2src.items()
    }

    if _args.infile:
        with open(_args.infile, mode='r') as fp:
            testSets = json.load(fp)
    else:
        xrc = xraygraphql.XrayClient(project=_args.project)
        testSets = xrc.ListTestSets()
        for mq, v in testSets.items():
            print(f'{mq} ', end='', file=sys.stderr)
            sys.stderr.flush()
            testSets[mq]['tests'] = xrc.GetTestsInTestSet(issueId=v['issueid'])
            sources = []
            for test, te in testSets[mq]['tests'].items():
                src_info = defs2src.get(te['unstructured'])
                if src_info is not None and src_info not in sources:
                    sources.append(src_info)
            testSets[mq]['sources'] = sources
        print('', end='', file=sys.stderr)

    if _args.outfile is not None:
        with open(_args.outfile, 'w') as fp:
            json.dump(testSets, fp, sort_keys=True, indent=4)
    else:
        for mq in sorted(testSets.keys(), key=lambda x: int(x.split('-')[-1])):
            entry = testSets[mq]
            print(f'{mq}: {entry["JIRA summary"]}')
            print('\t sources:', entry['sources'])
            print('\t tests:', list(entry['tests'].keys()))
