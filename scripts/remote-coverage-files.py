#!/usr/bin/env python3
'''
Simple script to retrieve coverage files
To reduce dependencies it uses subprocess to run
    - kubectl
    - scp

If this script is invoked and there are no coverage files
available on the nodes, no coverage directory is created
thus avoiding confusion and artifacts directory pollution
Calling scripts can be made simpler/more consistent,
i.e. no special casing for coverage builds.
'''

import subprocess
import sys
import yaml
import os
import tempfile
import shutil


def getCoverageFiles(**kwargs):
    '''
    Retrieve mayastor coverage files from nodes
    Relies on scp working without password

    returns the path to temporary directory into
        which coverage files have been copied
    '''
    results_dir = tempfile.mkdtemp()
    identity_file = kwargs['identity_file']
    result = subprocess.run([
        'kubectl', 'get', 'nodes', '-o', 'yaml'
    ],
        stdout=subprocess.PIPE)
    yamlstr = result.stdout.decode('utf-8')
    y = yaml.load(yamlstr, Loader=yaml.CLoader)

    nodes = []
    for item in y['items']:
        if 'openebs.io/engine' in item['metadata']['labels']:
            d = {}
            for addr in item['status']['addresses']:
                d[addr['type']] = addr['address']
            nodes.append(d)

    scp_cmds = [
        'scp',
        '-o',
        'StrictHostKeyChecking=no',
        '-o',
        'UserKnownHostsFile=/dev/null',
        '-r',
    ]
    if identity_file is not None and len(identity_file) != 0:
        scp_cmds.extend(['-i', identity_file])

    # DEVNULL = open(os.devnull, 'w')
    # We could use python scp from paramiko,
    # however using subprocess to invoke scp, means fewer dependencies
    for node in nodes:
        print(node)
        cmds = scp_cmds[:]
        cmds.extend([
            f'root@{node["InternalIP"]}:/var/local/mayastor-coverage',
            f'{results_dir}/{node["Hostname"]}'
        ])
        _ = subprocess.run(cmds,
                           # stdout=DEVNULL,
                           # stderr=DEVNULL
                           )
    return results_dir


def clearCoverageFiles(**kwargs):
    '''
    Clear mayastor coverage files from nodes
    Relies on ssh working without password

    '''
    identity_file = kwargs['identity_file']
    result = subprocess.run([
        'kubectl', 'get', 'nodes', '-o', 'yaml'
    ],
        stdout=subprocess.PIPE)
    yamlstr = result.stdout.decode('utf-8')
    y = yaml.load(yamlstr, Loader=yaml.CLoader)

    nodes = []
    for item in y['items']:
        if 'openebs.io/engine' in item['metadata']['labels']:
            d = {}
            for addr in item['status']['addresses']:
                d[addr['type']] = addr['address']
            nodes.append(d)

    scp_cmds = [
        'ssh',
        '-o',
        'StrictHostKeyChecking=no',
        '-o',
        'UserKnownHostsFile=/dev/null',
    ]
    if identity_file is not None and len(identity_file) != 0:
        scp_cmds.extend(['-i', identity_file])

    DEVNULL = open(os.devnull, 'w')
    # We could use python scp from paramiko,
    # however using subprocess to invoke scp, means fewer dependencies
    for node in nodes:
        print('Removing coverage files on ', node)
        cmds = scp_cmds[:]
        cmds.extend([
            f'root@{node["InternalIP"]}',
            'rm -v /var/local/mayastor-coverage/*'
        ])
        _ = subprocess.run(cmds,
                           stdout=DEVNULL,
                           stderr=DEVNULL
                           )


if __name__ == '__main__':
    from argparse import ArgumentParser
    import os.path

    _artefacts_dir = os.path.abspath(
        os.path.join(
            os.path.dirname(sys.argv[0]), '../artifacts'
        )
    )

    if not os.path.isdir(_artefacts_dir):
        _covpath = '/tmp/mayastor-coverage/data'
    else:
        _covpath = os.path.join(_artefacts_dir, 'coverage', 'data')

    parser = ArgumentParser()
    parser.add_argument('--path', dest='coverage_dir',
                        default=_covpath,
                        help='destination directory for coverage files'
                        )
    parser.add_argument('--identity_file', dest='identity_file',
                        default=None,
                        help='Selects the file from which the identity'
                        '(private key) for public key authentication is'
                        'read'
                        )
    parser.add_argument('--get', dest='get', action='store_true',
                        default=False,
                        help='Retrieve profile files from nodes in the cluster'
                        )
    parser.add_argument('--clear', dest='clear', action='store_true',
                        default=False,
                        help='Clear profile files on nodes in the cluster'
                        )

    _args = parser.parse_args()

    if _args.get:
        # Avoid creating coverage directory,
        # if no coverage files are present
        tmp_results_dir = getCoverageFiles(**vars(_args))
        if len(os.listdir(tmp_results_dir)) != 0:
            _args.coverage_dir = os.path.abspath(_args.coverage_dir)
            shutil.copytree(tmp_results_dir, _args.coverage_dir,
                            dirs_exist_ok=True)
            print(f'Coverage files stored in path {_args.coverage_dir}')
        shutil.rmtree(tmp_results_dir)
    elif _args.clear:
        clearCoverageFiles(**vars(_args))
