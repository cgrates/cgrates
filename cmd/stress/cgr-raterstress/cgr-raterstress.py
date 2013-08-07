# jsonclient.py
# A simple JSONRPC client library, created to work with Go servers
# Written by Stephen Day
# Modified by Bruce Eckel to work with both Python 2 & 3
import json, socket, itertools
from datetime import datetime

class JSONClient(object):

    def __init__(self, addr, codec=json):
        self._socket = socket.create_connection(addr)
        self._id_iter = itertools.count()
        self._codec = codec

    def _message(self, name, *params):
        return dict(id=next(self._id_iter),
                    params=list(params),
                    method=name)

    def call(self, name, *params):
        request = self._message(name, *params)
        id = request.get('id')
        msg = self._codec.dumps(request)
        self._socket.sendall(msg.encode())

        # This will actually have to loop if resp is bigger
        response = self._socket.recv(4096)
        response = self._codec.loads(response.decode())

        if response.get('id') != id:
            raise Exception("expected id=%s, received id=%s: %s"
                            %(id, response.get('id'),
                              response.get('error')))

        if response.get('error') is not None:
            raise Exception(response.get('error'))

        return response.get('result')

    def close(self):
        self._socket.close()


rpc =JSONClient(("127.0.0.1", 2012))

cd = {"Direction":"*out", "TOR":"call", "Tenant": "cgrates.org", "Subject": "1001", "Destination": "+49", "TimeStart": "2013-08-07T17:30:00Z", "TimeEnd": "2013-08-07T18:30:00Z"}

# alternative to the above
#s = socket.create_connection(("127.0.0.1", 2012))
#s.sendall(json.dumps({"id": 1, "method": "Responder.GetCost", "params": [cd]}))
#print(s.recv(4096))

i = 0
result = ""
for i in range(int(1e5) + 1):
    result = rpc.call("Responder.GetCost", cd)
print(i, result)
