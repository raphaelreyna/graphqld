#!/usr/bin/python3

import sys
import json

args = sys.argv

if 1 < len(args):
    if args[1] == "--cggi-fields":
        print("[\"charCount(strng: String!): CharCountResponse!\"]")
    if (args[1] == "--strng") & (len(args) == 3):
        strng = args[2]
        print(json.dumps({'input': strng, 'output': len(strng)}), end="")
