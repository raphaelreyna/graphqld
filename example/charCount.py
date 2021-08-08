#!/usr/bin/python3

import sys

args = sys.argv

if 1 < len(args):
    if args[1] == "--cggi-fields":
        print("[\"charCount(strng: String!): Int\"]")
    if (args[1] == "--strng") & (len(args) == 3):
        strng = args[2]
        print(len(strng), end="")

