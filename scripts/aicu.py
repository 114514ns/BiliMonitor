from curl_cffi import requests

import json

import time



r = requests.get("https://api.bilibili.com/x/web-interface/view/detail?bvid=BV1eiB7BKEVQ&isGaiaAvoided=true&w_rid=d04da50a34f33ca702d1553034fc5538&wts=1766473866")

print(r)

for i in range(1,19):
    r = requests.get(f"https://api.aicu.cc/api/v3/search/getreply?uid=504140200&pn={i}&ps=100&mode=0&keyword=", 
                 impersonate="chrome101",
                 ## proxies= { "http": "http://127.0.0.1:7890", "https": "http://127.0.0.1:7890", }
                 )

    obj = json.loads(r.text)
    print(f"page={i}")
    print()
    print()
    for item in obj['data']['replies']:
        print(item['message'])
        
    time.sleep(3)