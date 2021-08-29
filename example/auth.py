#!/usr/bin/python3

import os
import json
import base64
import sys

if "HTTP_AUTHORIZATION" not in os.environ:
    print(json.dumps({"user": {}, "loggedIn": False}), end="")
    sys.exit(0)

authParts = os.environ["HTTP_AUTHORIZATION"].split()
if len(authParts) < 1:
    print(json.dumps({"user": {}, "loggedIn": False}), end="")
    sys.exit(0)

credentials = base64.b64decode(authParts[1].encode('ascii')).decode('ascii').split(':')
print(json.dumps({"user": {"name": credentials[0]}, "loggedIn": True}), end="")