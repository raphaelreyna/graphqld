#!/usr/bin/python3

import sys
import json

args = sys.argv

if 1 < len(args):
    if args[1] == "--cggi-fields":
        print("[\"isEven: IsEvenResponse!\"]")
else:
    source = json.load(sys.stdin)
    isEven  = source['count'] % 2 == 0
    print(json.dumps({'response': isEven}), end="")
