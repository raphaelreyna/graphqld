#!/usr/bin/python3

import sys
import json
import os

if __name__ == '__main__':
    args = sys.argv
    if 1 < len(args):
        if args[1] == "--graphqld-fields":
            print("[\"python(ssss: PythonInput): String!\"]")
        else:
            inputJSON = args[2]
            input = json.loads(inputJSON)
            python = "python"

            try:
                ctxFile = open("/dev/fd/3", "r")
                ctx = json.load(ctxFile)
                ctxFile.close()
                if ctx["loggedIn"] is True:
                    python += f' (I know your name is really {ctx["user"]["name"]})'
            except FileNotFoundError:
                None

            if "LevelTwo" in input:
                if input["LevelTwo"]["IncludeVersion"]:
                    python = "python3"

            print(f'Hi {input["Name"]}, this is {python}!', end="")