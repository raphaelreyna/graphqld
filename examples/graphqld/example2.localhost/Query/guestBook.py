#!/usr/bin/python3

import sys
import json

args = sys.argv

if 1 < len(args):
    if args[1] == "--graphqld-fields":
        print("[\"guestBook(e: Status): [String]\"]")
else:
    with open("guestbook.txt", "r+") as f:
        guests = f.read().splitlines() 
        print(json.dumps(guests), end="")
