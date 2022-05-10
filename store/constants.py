import sys
import os

isUnix = sys.platform in ["linux", "darwin"]
processID = os.getpid()
