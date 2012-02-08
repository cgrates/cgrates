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


rpc =JSONClient(("127.0.0.1", 5090))

cd = {"Tor":0, "CstmId": "vdf", "Subject": "rif", "DestinationPrefix": "0256", "TimeStart": "2012-02-02T17:30:00Z", "TimeEnd": "2012-02-02T18:30:00Z"}

# alternative to the above
s = socket.create_connection(("127.0.0.1", 5090))
s.sendall(json.dumps(({"id": 1, "method": "Responder.Get", "params": [cd]})))
print s.recv(4096)

i = 0
result = ""
for i in xrange(int(1e4) + 1):
    result = rpc.call("Responder.Get", cd)
print i, result
