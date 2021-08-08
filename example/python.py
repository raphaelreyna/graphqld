#!/usr/bin/python3

import sys

if __name__ == '__main__':
    args = sys.argv

    if 1 < len(args):
        if args[1] == "--cggi-fields":
            print("[\"python: String!\"]")
    else:
        print("Hello from python!", end="")