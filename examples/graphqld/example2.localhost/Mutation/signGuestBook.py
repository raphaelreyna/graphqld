#!/usr/bin/python3

import sys
import json

args = sys.argv

if 1 < len(args):
    if args[1] == "--graphqld-fields":
        print("[\"signGuestBook(name: String!): SignGuestBookResponse!\"]")
    if args[1] == "--name" and len(args) is 3:
        file = open("guestbook.txt", "a+")
        file.write(args[2])
        file.write("\n")
        file.close
        print(json.dumps({"Name": args[2]}), end="")