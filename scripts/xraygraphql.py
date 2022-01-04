#!/usr/bin/env python3
'''
This module defines a class which uses interfaces with the Xray GraphQL
interface to enumerate Test, TestSets, TestPlans, TestExecutions and
TestRuns
'''
from datetime import datetime
import json
import os
import requests
import sys
from types import SimpleNamespace
import yaml

trace = False

_client_id = None
_client_secret = None
_auth_url = 'https://xray.cloud.xpand-it.com/api/v1/authenticate'

_gql_url = 'https://xray.cloud.xpand-it.com/api/v2/graphql'


class XrayClient:
    '''
    The XrayClient object makes Xray GraphQL calls to retrieve Test object info

    Args:
        project (str) : project name used in for jql e.g. 'MQ'
        graphql_url (str) : graphql server url, defaulted
        auth_url (str) : authentication REST server url, defaulted
        client_id (str) : client ID, defaulted
        client_secret (str) : client secretd, defaulted

    '''

    def __init__(self,
                 project,
                 graphql_url=_gql_url,
                 auth_url=_auth_url,
                 client_id=_client_id,
                 client_secret=_client_secret
                 ):
        if client_id is None or client_secret is None:
            try:
                yf = os.getenv('HOME') + '/.xrayclient.yaml'
                with open(yf, mode='r') as fp:
                    cl = yaml.safe_load(fp.read())
                    client_id = cl['client_id']
                    client_secret = cl['client_secret']
                    print(f'Using client credentials from {yf}')
            except FileNotFoundError:
                client_id = os.getenv('XRAY_CLIENT_ID', None)
                client_secret = os.getenv('XRAY_CLIENT_SECRET', None)
        if client_id is None:
            raise Exception(
                'client_id is None,'
                'please define environment variable XRAY_CLIENT_ID')
        if client_secret is None:
            raise Exception(
                'client_secret is None,'
                'please define environment variable XRAY_CLIENT_SECRET')

        self.headers = {
            'Content-type': 'application/json',
            'Accept': 'text/plain'
        }
        self.auth_url = auth_url
        self.graphql_url = graphql_url
        self.client_id = client_id
        self.client_secret = client_secret
        self.project = project

    def SetProject(self, project):
        ''' Set the project name for queries

        Args:
            project (string) : project name
        '''
        self.project = project

    def _post_json_data(self, url, data):
        resp = requests.post(url, data=json.dumps(data), headers=self.headers)
        return resp

    def _post(self, url, data):
        resp = requests.post(url, data=data, headers=self.headers)
        return resp

    def _authenticate(self, again=False):
        if not again and 'Authorization' in self.headers.keys():
            return
        url = self.auth_url
        payload = {
            'client_id': self.client_id,
            'client_secret': self.client_secret
        }
        headers = {"Content-Type": "application/json"}
        resp = requests.post(url, payload, headers)
        self.bearer_token = resp.json()
        self.headers.update({
            'Authorization': f'Bearer {self.bearer_token}'
        })

    def Query(self, **kw_args):
        """ Make graphql query

        Args:
            **kw_args
                The keyword arguments for the query
                    "query" : <graphql query>,
                    "variables" : <dict | None>,
                    "operationName": <string>,
                    "trace": <True|False|None>
                    "variables", "operationName" and "trace" can be absent
        """
        self._authenticate()
        query = kw_args.get('query')
        if query is None:
            raise Exception('Empty query')
        variables = kw_args.get('variables')
        if variables is not None:
            variables = json.dumps(variables)

        post_data = json.dumps({
            'query': f'query {query}',
            'variables': variables,
            'operationName': kw_args.get('operationName')
        })
        if kw_args.get('trace'):
            print(f'TRACE: post_data = {post_data}', file=sys.stderr)
        resp = self._post(self.graphql_url, post_data)
        return resp.json()

    def _collect(self, query=None, variables=None, operationName=None):
        """ collect results from graphql queries which return list fragments
        limitations:
                1. handles only a single set of results (no nesting)
                2. handles results returned upto 1 level down

        Args:
            query (str) : GraphQL query string
            variables (dict): dict for variables used in the query
            operationName (str):
        """
        self._authenticate()
        if query is None:
            return None
        ctxt = SimpleNamespace(
            results=[],
            varz=variables,
            total=0,
            more=True,
            abort=False
        )

        def _processResults(ctx, datum):
            """ process the response, collect results and update the context
            """
            for j, v in datum.items():
                if j == 'results':
                    count = len(v)
                    # FIXME: if we get a 0 sized results array,
                    # we do not really know how to handle that correctly.
                    # For now abort the iteration.
                    if count == 0:
                        ctx.more = False
                        ctx.abort = True
                        if trace:
                            print(
                                f'TRACE: got: {count} results',
                                f' total={ctx.total}'
                                f' len results={len(ctx.results)}',
                                file=sys.stderr
                            )
                        break
                    # update the results list
                    ctx.results.extend(v)
                    # update the variables to get the next N entries
                    ctx.varz['start'] += len(v)
                elif j == 'total':
                    if trace:
                        print(f'TRACE: got: {j} = {v}', file=sys.stderr)
                    # update the totals as returned
                    ctx.total = v
                    if ctx.total == 0:
                        ctx.more = False
                        break
                elif j == 'limit':
                    if trace:
                        print(f'TRACE: got: {j} = {v}', file=sys.stderr)
                    # update the limit as returned
                    if v < ctx.varz['limit']:
                        ctx.varz['limit'] = v
                else:
                    # ignore all other entries
                    if trace:
                        print(f'TRACE: ignoring: {j} = {v}', file=sys.stderr)

        # synthesized  do-while loop
        ctxt.more = True
        while ctxt.more:
            if trace:
                print(f'TRACE: query variables {ctxt.varz}', file=sys.stderr)
            post_data = json.dumps({
                'query': f'query {query}',
                'variables': json.dumps(ctxt.varz),
                'operationName': operationName,
            })
            resp = self._post(self.graphql_url, post_data)
            if resp.status_code != 200:
                err = f'request failed {resp.status_code}, {resp.reason}'
                print(f'TRACE: {err}', file=sys.stderr)
                raise(Exception(err))
            x = resp.json()['data']
            for k in x:
                y = x[k]
                if 'results' in y.keys():
                    _processResults(ctxt, y)
                else:
                    # handle results returned one level below the top
                    for j, v in y.items():
                        if isinstance(v, dict):
                            if 'results' in v.keys():
                                _processResults(ctxt, v)
            ctxt.more = (ctxt.more
                         and len(ctxt.results) < ctxt.total
                         and not ctxt.abort)

        if trace:
            print(
                f'TRACE: total={ctxt.total},'
                f' len(results)={len(ctxt.results)}',
                f' abort={ctxt.abort}',
                file=sys.stderr
            )
        return ctxt.results

    def GetTest(self, issueId):
        ''' Get tests

        Returns:
            a dictionary of tests
        '''
        query = '''
        GetTest($issueId: String!) {
        getTest(issueId: $issueId) {
                    jira(fields: ["key", "summary", "status"])
                    unstructured
                    testSets(limit: 10, start: 0) {
                        results {
                            issueId
                            jira(fields: ["key"])
                        }
                    }
                }
        }
        '''
        results = self.Query(**{
            'query': query,
            'variables': {'issueId': issueId}
        })
        return results

    def ListTests(self):
        ''' List tests

        Returns:
            a dictionary of tests, keys on Jira Key
        '''
        # FIXME: number of testSets and testPlans is fixed at 10
        # we may want to do something about that in the future.
        query = '''
        ListTests($limit: Int!, $start: Int!, $jql: String!) {
                getTests(
                    jql: $jql, limit: $limit, start: $start
                ) {
                total
                start
                limit
                results {
                    issueId
                    jira(fields: ["key", "summary", "status"])
                    unstructured
                    testSets(limit: 10, start:0) {
                        results {
                            jira(fields: ["key"])
                        }
                    }
                    testPlans(limit: 10, start:0) {
                        results {
                            jira(fields: ["key"])
                        }
                    }
# sample testRuns used to check if the test is being run
                    testRuns(limit: 10, start:0) {
                        results {
                            finishedOn
                        }
                    }
                }
            }
        }
        '''
        variables = {
            'jql': f"project = {self.project}",
            'limit': 50,
            'start': 0
        }

        results = self._collect(
            query=query, variables=variables)
        tests = {x['jira']['key']: x for x in results}
        return tests

    def ListTestPlans(self):
        ''' List test plans

        Returns:
            a dictionary of test plans, keys on Jira Key
        '''
        query = '''
        ListTestPlans($limit: Int!, $start: Int!, $jql: String!) {
                getTestPlans(
                    jql: $jql, limit: $limit, start: $start
                ) {
                total
                start
                limit
                results {
                    issueId
                    jira(fields: ["key", "summary", "description"])
                }
            }
        }
        '''
        variables = {
            'jql': f"project = {self.project}",
            'limit': 50,
            'start': 0
        }

        results = self._collect(
            query=query, variables=variables)
        testPlans = {x['jira']['key']: x for x in results}
        return testPlans

    def GetTestsInTestPlan(self, issueId=None, jiraKey=None):
        ''' Get the list of tests in a testplan

        Args:
            issueId (str): test plan issue id
            jiraKey (str): jira key for the test plan

        only one of issueId or jiraKey is required,
        issueId is preferred over jiraKey


        Returns:
            a dictionary of tests in the test plan, keyed on Jira Key
        '''
        if issueId is None:
            if jiraKey is not None:
                tps = self.ListTestPlans()
                if jiraKey not in tps:
                    raise(Exception('testplan {} not found'.format(jiraKey)))
                issueId = tps[jiraKey]
                if isinstance(issueId, dict):
                    issueId = issueId['issueId']
            else:
                raise(Exception('issueId or jira key required'))

        query = '''
        GetTestsInTestPlan($limit: Int!, $start: Int!, $issueId: String!) {
            getTestPlan(issueId: $issueId) {
                issueId
                tests(limit: $limit, start: $start) {
                    total
                    start
                    limit
                    results {
                        issueId
                        unstructured
                        jira(fields: ["key", "summary"])
                    }
                }
            }
        }
        '''
        results = self._collect(
            query=query,
            variables={'limit': 50, 'start': 0, 'issueId': issueId}
        )
        tests = {x['jira']['key']: x for x in results}
        return tests

    def GetTestExecutionsInTestPlan(self, issueId=None, jiraKey=None):
        ''' Get the list of test executions in a testplan

        Args:
            issueId (str): test plan issue id
            jiraKey (str): jira key for the test plan

        only one of issueId or jiraKey is required,
        issueId is preferred over jiraKey


        Returns:
            a dictionary of test executions keyed on Jira Key
        '''
        if issueId is None:
            if jiraKey is not None:
                tps = self.ListTestPlans()
                if jiraKey not in tps:
                    raise(Exception('testplan {} not found'.format(jiraKey)))
                issueId = tps[jiraKey]
                if isinstance(issueId, dict):
                    issueId = issueId['issueId']
            else:
                raise(Exception('issueId or jira key required'))

        query = '''
        GetTestExecutionsInTestPlan(
            $limit: Int!, $start: Int!, $issueId: String!) {
            getTestPlan(issueId: $issueId) {
                issueId
                testExecutions(limit: $limit, start: $start) {
                    total
                    start
                    limit
                    results {
                        issueId
                        jira(fields: ["key", "summary"])
                        lastModified
                    }
                }
            }
        }
        '''
        results = self._collect(
            query=query,
            variables={'limit': 50, 'start': 0, 'issueId': issueId}
        )
        tests = {x['jira']['key']: x for x in results}
        return tests

    def ListTestSets(self):
        ''' List test sets

        Returns:
            a dictionary of test sets, keyed on Jira Key
        '''
        query = '''
        ListTestSets($limit: Int!, $start: Int!, $jql: String!) {
            getTestSets(jql: $jql, limit: $limit, start: $start) {
                    total
                    start
                    limit
                    results {
                        issueId
                        jira(fields: ["key", "summary"])
                    }
                }
        }
        '''
        variables = {
            'jql': f"project = {self.project}",
            'limit': 50,
            'start': 0
        }

        results = self._collect(
            query=query, variables=variables)
        testSets = {x['jira']['key']: {
            'issueid': x['issueId'],
            'JIRA summary': x['jira']['summary']
        }for x in results}
        return testSets

    def GetTestsInTestSet(self, issueId=None, jiraKey=None):
        ''' Get the list of tests in a testset

        Args:
            issueId (str): test set issue id
            jiraKey (str): jira key for the test set

        only one of issueId or jiraKey is required,
        issueId is preferred over jiraKey


        Returns:
            a dictionary of dictionaries:
                dictionary of test sets keyed on test set Jira Key,
                each entry is a dictionary of tests in the test set
                keyed on test Jira Key
        '''
        if issueId is None:
            if jiraKey is not None:
                tsets = self.ListTestSets()
                if jiraKey not in tsets:
                    raise(Exception('testSet {} not found'.format(jiraKey)))
                issueId = tsets[jiraKey]
                if isinstance(issueId, dict):
                    issueId = issueId['issueid']
            else:
                raise(Exception('issueId or jira key required'))

        query = '''
        GetTestsInTestSet($limit: Int!, $start: Int!, $issueId: String!) {
            getTestSet(issueId: $issueId) {
                issueId
                tests(limit: $limit, start: $start) {
                    total
                    start
                    limit
                    results {
                        issueId
                        unstructured
                        jira(fields: ["key", "summary"])
                    }
                }
            }
        }
        '''
        results = self._collect(
            query=query,
            variables={'limit': 50, 'start': 0, 'issueId': issueId}
        )
        tests = {x['jira']['key']: x for x in results}
        return tests

    def GetTestsInAllTestSets(self):
        ''' Get the list of tests in all testsets

        Returns:
            a dictionary of dictionaries:
                dictionary of test sets keyed on test set Jira Key,
                each entry is a dictionary of tests in the test set
                keyed on test Jira Key
        '''
        testSetTests = {}
        testSets = self.ListTestSets()
        for jiraKey, issueId in testSets.items():
            if isinstance(issueId, dict):
                issueId = issueId['issueid']
            testSetTests[jiraKey] = self.GetTestsInTestSet(issueId=issueId)
        return testSetTests

    def ListTestExecutions(self):
        ''' List test executions

        Returns:
            a dictionary of test executions keyed on Jira Key
        '''
        query = '''
        ListTestExecutions($limit: Int!, $start: Int!, $jql: String!){
            getTestExecutions(
                jql: $jql, limit: $limit, start: $start
                ) {
                total
                start
                limit
                results {
                    issueId
                    jira(fields: ["key", "summary"])
                }
            }
        }
        '''
        variables = {
            'jql': f"project = {self.project}",
            'limit': 50,
            'start': 0
        }

        results = self._collect(
            query=query, variables=variables)
        testExecutions = {x['jira']['key']: x for x in results}
        return testExecutions

    def GetTestExecutionRuns(self, issueId=None, jiraKey=None):
        ''' Get the list of test runs in a test execution

        Args:
            issueId (str): test execution issue id
            jiraKey (str): jira key for the test plan

        only one of issueId or jiraKey is required,
        issueId is preferred over jiraKey


        Returns:
            list of test runs in a test execution
        '''
        if issueId is None:
            if jiraKey is not None:
                tsets = self.ListTestExecutions()
                if jiraKey not in tsets:
                    raise(Exception(f'test execution {jiraKey} not found'))
                issueId = tsets[jiraKey]
                if isinstance(issueId, dict):
                    issueId = issueId['issueId']
            else:
                raise(Exception('issueId or jira key required'))

        query = '''
        GetTestExecutionRuns($limit: Int!, $start: Int!, $issueId: String!) {
            getTestExecution(issueId: $issueId) {
                issueId
                testRuns(limit: $limit, start: $start) {
                    total
                    start
                    limit
                    results {
                        id
                        unstructured
#                        startedOn
                        finishedOn
                        status {
                            name
#                            description
#                            final
#                            color
                        }
                        test {
                            jira(fields: ["key", "summary"])
                        }
                        results {
                            log
                        }
                    }
                }
            }
        }
        '''
        testRuns = self._collect(
            query=query,
            variables={'limit': 50, 'start': 0, 'issueId': issueId})

        return testRuns

    def GetTestRuns(self, issueId=None, jiraKey=None):
        ''' Get the list of test runs

        Args:
            issueId (str): test issue id
            jiraKey (str): jira key for the test

        only one of issueId or jiraKey is required,
        issueId is preferred over jiraKey


        Returns:
            list of test runs in a test execution
        '''
        if issueId is None:
            if jiraKey is not None:
                tst = self.ListTests()
                if jiraKey not in tst:
                    raise(Exception(f'test execution {jiraKey} not found'))
                issueId = tst[jiraKey]
                if isinstance(issueId, dict):
                    issueId = issueId['issueId']
            else:
                raise(Exception('issueId or jira key required'))

        query = '''
        GetTestRuns($limit: Int!, $start: Int!, $issueId: String!) {
            getTest(issueId: $issueId) {
                issueId
                testRuns(limit: $limit, start: $start) {
                    total
                    start
                    limit
                    results {
                        id
                        unstructured
#                        startedOn
                        finishedOn
                        status {
                            name
                            description
#                            final
#                            color
                        }
                        test {
                            jira(fields: ["key", "summary"])
                        }
                        results {
                            log
                        }
                    }
                }
            }
        }
        '''
        testRuns = self._collect(
            query=query,
            variables={'limit': 50, 'start': 0, 'issueId': issueId})
        return testRuns

    def GetTestRunResult(self, id=None):
        query2 = '''
        GetTestRunResult($id: String!) {
            getTestRunById(id: $id) {
                results {
                    log
                }
            }
        }
        '''

        return self.Query(**{
            'query': query2,
            'variables':  {'id': id}
        })
