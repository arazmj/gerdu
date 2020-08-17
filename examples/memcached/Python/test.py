from pymemcache.client import base

if __name__ == '__main__':
    client = base.Client(('localhost', 11211))
    client.set('Hello', 'World')
    value = client.get('Hello').decode("utf-8")
    print('Hello =', value)

