#!/usr/bin/env python

import os
import time

start=1331452800
end=int(time.time())

for t in range(start, end, 3600):
    date = time.strftime('%F-', time.localtime(t))
    hour = int(time.strftime('%H', time.localtime(t)))
    fnbase = date + str(hour)
    fn = fnbase + '.json.gz'
    if not os.access(fn, os.R_OK):
        print fnbase
