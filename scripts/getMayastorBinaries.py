#!/usr/bin/env python
'''
Simple dumb script to retrieve primary binaries from docker image files.

Use to retrieve mayastor binaries for code coverage analysis
'''

import json
import os.path
import subprocess
import sys


def copyOneContainerExecutable(containerId, outputdir):
    op = subprocess.check_output(['docker', 'inspect', containerId])
    jsondata = op.decode('utf-8')
    cdata = json.loads(jsondata)
    if len(cdata) == 1:
        data = cdata[0]
        exe = data['Path']
        if exe == 'tini':
            args = data['Args']
            exe = f'/bin/{args[1]}'

        print(f'copying {exe} to {outputdir}')
        try:
            op = subprocess.check_output([
                'docker',
                'cp',
                f'{containerId}:{exe}',
                f'{outputdir}',
            ])
        except subprocess.CalledProcessError:
            print(f'failed to copy {exe} to {outputdir}')
    else:
        print('Unexpected: inspect returned more than one container')


def copyContainerExecutables(**kwargs):
    '''
    Copy binary files from containers to specified location.
    Simple and dumb, uses subprocess and docker to achieve objective
    '''
    registry = kwargs['registry']
    if registry != "":
        registry += '/'

    for img in kwargs['images']:
        print(f'Examining {img}....')
        image = f'{registry}mayadata/{img}:{kwargs["tag"]}'
        try:
            op = subprocess.check_output(['docker', 'create', image])
        except subprocess.CalledProcessError as cpe:
            print(f'Failed to create container {image}', cpe)
            continue

        containerId = op.decode('utf-8')
        containerId = containerId.strip()

        try:
            copyOneContainerExecutable(containerId, kwargs['outputdir'])
        except subprocess.CalledProcessError as cpe:
            print("Failed to copying executable", cpe)

        op = subprocess.check_output(['docker', 'rm', containerId])


if __name__ == '__main__':
    from argparse import ArgumentParser

    parser = ArgumentParser()
    parser.add_argument('--tag', dest='tag',
                        default=None,
                        required=True,
                        help='build tag'
                        )
    parser.add_argument('--outputdir', dest='outputdir',
                        default=None,
                        help='location to copy binaries'
                        )
    parser.add_argument('--registry', dest='registry',
                        default="",
                        help='registry'
                        )
    parser.add_argument('images', nargs='*')
    _args = parser.parse_args()
    if _args.outputdir is None:
        _artefacts_dir = os.path.abspath(
            os.path.join(
                os.path.dirname(sys.argv[0]), '../artifacts'
            )
        )

        if not os.path.isdir(_artefacts_dir):
            _args.outputdir = f'/tmp/mayastor/binaries/{_args.tag}'
        else:
            _args.outputdir = os.path.join(
                _artefacts_dir, 'binaries', _args.tag
            )

    # FIXME: remove the need for a hardcoded list.
    # TODO:
    #   1 - generate install yamls using the install bundle
    #   2 - parse the install yamls and record the set of mayastor images
    #   3 - use that set as the list of images
    if len(_args.images) == 0:
        _args.images = [
            "mayastor-cov",
            "mcp-core-cov",
            "mcp-csi-controller-cov",
            "mcp-msp-operator-cov",
            "mcp-rest-cov"
        ]

    if not os.path.exists(_args.outputdir):
        os.makedirs(_args.outputdir)

    copyContainerExecutables(**vars(_args))
