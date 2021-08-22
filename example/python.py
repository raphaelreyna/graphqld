#!/usr/bin/python3

import sys

if __name__ == '__main__':
    args = sys.argv

    if 1 < len(args):
        if args[1] == "--cggi-fields":
            print("[\"python(ssss: PythonInput): String!\"]")
    else:
        print("Test-Header: test-value\n\n", end="")
        print("Hello from python!", end="")