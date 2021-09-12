#!/usr/bin/python3

import sys
import json

args = sys.argv

if 1 < len(args):
    if args[1] == "--graphqld-fields":
        print("[\"charCount(string: String!): CharCountResponse!\"]")
    if (args[1] == "--string") & (len(args) == 3):
        strng = args[2]
        print(json.dumps({'string': strng, 'count': len(strng)}), end="")
