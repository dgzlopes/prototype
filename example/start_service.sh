#!/bin/sh
python3 /code/service.py &
/prototype -d -cluster default -service service -tags=env:development -prototype-url http://prototype:10000