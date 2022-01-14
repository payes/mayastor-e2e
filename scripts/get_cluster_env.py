#!/usr/bin/env python
"""
Retrieve test cluster environment and check that all worker
nodes have the same environment
optionally:
    - print the values to screen
    - write xray test environments string to a file
"""

from datetime import datetime
import subprocess
import time
import yaml

DISTRO = "distribution"
VERSION = "version"

XRAY_K8SMINOR = 'k8s_minor'
XRAY_K8SPATCH = 'k8s_patch'
XRAY_KERNEL = 'os_kernel'
XRAY_OSDISTRO = 'os_dist'
XRAY_OSVER = 'os_release'
XRAY_PLATFORM = 'platform'

# load from a configuration file.
translate = {
    'kernel': {
        '5.13.0-27-generic': '5.13.0-27-generic'
    },
    'os': {
        'Ubuntu 20.04.3 LTS': {
            DISTRO: 'ubuntu',
            VERSION: '20.04.3_lts',
        }
    },
}

def testenvs(node):
    """
    xray environment string generator for a k8s node
    """
    patch = node['status']['nodeInfo']['kubeletVersion'].replace(
        ' ', '_')[1:]
    # k8s patch
    yield XRAY_K8SPATCH, patch

    # k8s major.minor
    yield XRAY_K8SMINOR, '.'.join(patch.split('.')[:-1])

    # kernel value may be translated to a token suitable for xray test
    # environment
    try:
        kernel = translate['kernel'][
            node['status']['nodeInfo']['kernelVersion']
        ]
    except KeyError:
        kernel = node['status']['nodeInfo']['kernelVersion'].replace(
            ' ', '_')
    yield XRAY_KERNEL, kernel

    # os distribution and version values are always translated to tokens
    # suitable for xray test environments
    for prefix, key in [(XRAY_OSDISTRO, DISTRO), (XRAY_OSVER, VERSION)]:
        try:
            value = translate['os'][node['status']['nodeInfo']['osImage']][key]
            yield prefix, value
        except KeyError:
            pass


def get_worker_node_envs(args):
    """
    returns the set of environment strings for worker nodes in the k8s cluster
    """
    result = subprocess.run([
        'kubectl', 'get', 'nodes', '-o', 'yaml'],
        stdout=subprocess.PIPE, check=True)
    yamlstr = result.stdout.decode('utf-8')
    yml = yaml.load(yamlstr, Loader=yaml.CLoader)
    node_envs = set()
    node_env_dicts = []
    for node in yml['items']:
        if 'node-role.kubernetes.io/master' in node['metadata']['labels']:
            continue
        if args.verbose:
            print(node['metadata']['name'],
                  ':\n',
                  '\tkubeletVersion: ',
                  node['status']['nodeInfo']['kubeletVersion'],
                  '\tkernelVersion: ',
                  node['status']['nodeInfo']['kernelVersion'],
                  '\tosImage: ',
                  node['status']['nodeInfo']['osImage']
                  )

        envs_dict = {v[0]:v[1] for v in testenvs(node)}
        node_env_dicts.append(envs_dict)
        xray_envs = [f'{k}={v}' for k,v in envs_dict.items()]
        node_envs.add(';'.join(sorted(xray_envs[:18])))
    if args.verbose:
        print('----------')
    return node_envs, node_env_dicts


def main(args):
    """
    Query k8s worker nodes settings to determine environment values
    and optionally store the values in a file.

    throws RuntimeError if worker node enviroments are different.
    """

    def check_and_save(envs, args):
        if len(envs) != 1:
            return False

        envstring = envs.pop()
        if args.platform is not None:
            if args.platform.startswith('plat_'):
                args.platform = args.platform[5:]
            envstring = f'platform={args.platform};{envstring}'
            node_env_dicts[0]['platform'] = args.platform

        if args.verbose:
            print('\nxray_test_environments: ', envstring)
            print(yaml.dump(node_env_dicts[0]))

        if args.xray_outfile is not None:
            with open(args.xray_outfile, "w", encoding='UTF-8') as f_p:
                f_p.write(envstring)

        if args.yaml_outfile is not None:
            with open(args.yaml_outfile, "w", encoding='UTF-8') as f_p:
                yaml.dump(node_env_dicts[0], f_p)

        return True

    for _ in range(args.retry):
        print(datetime.now())
        node_envs, node_env_dicts = get_worker_node_envs(args)

        if check_and_save(node_envs, args):
            return

        time.sleep(args.sleeptime)

    raise RuntimeError("All worker nodes do not have same environment")


if __name__ == '__main__':
    from argparse import ArgumentParser
    parser = ArgumentParser()
    parser.add_argument('-v', dest='verbose', action='store_true',
                        default=None,
                        help='verbose')
    parser.add_argument('--platform', dest='platform', default=None)
    parser.add_argument('--oxray', dest='xray_outfile', default=None,
                        help='path to the output file')
    parser.add_argument('--oyaml', dest='yaml_outfile', default=None,
                        help='path to output yaml file')
    parser.add_argument('--retry', dest='retry', type=int, default=1,
                        help='retry count')
    parser.add_argument('--sleeptime', dest='sleeptime', type=int,
                        help='retry timeout seconds')

    main(parser.parse_args())
