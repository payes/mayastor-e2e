#!/usr/bin/env python
"""
Load a test list definition yaml file and
output  the list of tests for a specified profile
The default delimiter is a space
Optionally
    - change the delimiter
    - sort the list of tests alphbetically
    - sort the list of tests in reverse order of execution time
    - return a list where each test is bracketed by install and uninstall
    - return a list which starts with install and ends with uninstall
"""

import os
import yaml


def main(args):
    """
    The real main function
    """
    if args.install_uninstall:
        if args.install or args.uninstall:
            raise RuntimeError(
                'Incompatible options --iu with --install or --uninstall')
    with open(args.listdef, mode='r', encoding='UTF-8') as f_p:
        contents = f_p.read()
        list_defs = yaml.safe_load(contents)
        durations = list_defs['metadata']['recorded_durations']
        tests = list_defs['testprofiles'][args.profile]

        if args.sort_alpha:
            tests = sorted(tests)
        if args.sort_duration:
            tests = sorted(tests, key=lambda x:durations.get(x,0), reverse=True)

        if args.install_uninstall:
            tests = [f'install,{tst},uninstall' for tst in tests]
        if args.install:
            tests.insert(0, 'install')
        if args.uninstall:
            tests.append('uninstall')

        if args.outputfile is None:
            print(args.separator.join(tests))
        else:
            with open(args.outputfile, "w", encoding="UTF-8") as fout:
                print(args.separator.join(tests), file=fout)


if __name__ == '__main__':
    from argparse import ArgumentParser
    parser = ArgumentParser()
    parser.add_argument('--lists', dest='listdef',
                        default=os.path.abspath(os.path.join(os.path.dirname(
                            os.path.realpath(__file__)),
                            '../configurations/testlists.yaml')),
                        help='list definitions')
    parser.add_argument('--profile', dest='profile', default=None,
                        help='profile')

    parser.add_argument('--install', dest='install', action='store_true',
                        default=None,
                        help='Add install before all tests')
    parser.add_argument('--uninstall', dest='uninstall', action='store_true',
                        default=None,
                        help='Add uninstall before all tests')
    parser.add_argument('--iu', dest='install_uninstall', action='store_true',
                        default=None,
                        help='Each test is preceded with install and followed with an uninstall')
    parser.add_argument('--separator', dest='separator', default=' ',
                        help='seperator')

    parser.add_argument('--sort_alpha', dest='sort_alpha', action='store_true',
                        default=False,
                        help='sort lists alphabetically')
    parser.add_argument('--sort_duration', dest='sort_duration', action='store_true',
                        default=False,
                        help='sort the list in order of execution time high to low')

    parser.add_argument('--outputfile', dest='outputfile', default=None,
                        help='outputfile')



    main(parser.parse_args())
