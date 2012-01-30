# jsonclient.py
# A simple JSONRPC client library, created to work with Go servers
# Written by Stephen Day
# Modified by Bruce Eckel to work with both Python 2 & 3
import json, socket, itertools

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

# alternative to the above
s = socket.create_connection(("127.0.0.1", 5090))
s.sendall(json.dumps(({"id": 1, "method": "Responder.Get", "params": ["test"]})))
print s.recv(4096)

i = 0
result = ""
for i in xrange(5 * int(10e4) + 1):
    result = rpc.call("Responder.Get", "test")
print i, result
