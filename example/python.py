#!/usr/bin/python3

import sys
import json

if __name__ == '__main__':
    args = sys.argv
    if 1 < len(args):
        if args[1] == "--cggi-fields":
            print("[\"python(ssss: PythonInput): String!\"]")
        else:
            inputJSON = args[2]
            input = json.loads(inputJSON)
            print(f'Hi {input["Name"]}, this is python!', end="")