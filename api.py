import requests
import json

url = "http://localhost:8080/api"

def newKey():
    return requests.get(url+"/new").json()['key']
    
def nextPrompt(key):
    req = {
        'key': key,
    }
    return requests.post(url+"/next", data=json.dumps(req)).json()

def answerPrompt(key, ans):
    req = {
        'key': key,
        'answer': ans,
    }
    requests.post(url+"/answer", data=json.dumps(req))

def skip(key):
    req = {
        'key': key,
    }
    requests.post(url+"/skip", data=json.dumps(req))


key = newKey()
n = nextPrompt(key)
print("New Madlib: {0}".format(n['title']))
while True:
    if n['done']:
        print(n['madlib'])
        break
    print("Give me a {0}".format(n['prompt']))
    s = raw_input("> ").strip()
    if s == 'skip':
        print("Skipping...")
        skip(key)
        n = nextPrompt(key)
        print("New Madlib: {0}".format(n['title']))
    elif s == 'exit':
        print('Bye!')
        break
    elif s:
        answerPrompt(key, s)
        n = nextPrompt(key)
        
