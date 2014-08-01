package service

import (
    "net/http"
    "net/url"
)

//指定代理ip
func GetByProxy(url_addr, proxy_addr string) (*http.Response, error) {
    request, _ := http.NewRequest("GET", url_addr, nil)
    request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
    // request.Header.Add("Accept-Encoding", "gzip,deflate,sdch")
    request.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,en;q=0.6")
    request.Header.Add("Connection", "keep-alive")
    request.Header.Add("DNT", "1")
    request.Header.Add("Host", "www.baidu.com")
    request.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/31.0.1650.63 Chrome/31.0.1650.63 Safari/537.36")
    proxy_addr = "http://" + proxy_addr + "/"
    proxy, err := url.Parse(proxy_addr)
    if err != nil {
        return nil, err
    }
    client := http.Client{
        Transport: &http.Transport{
            Proxy: http.ProxyURL(proxy),
        },
    }
    return client.Do(request)
}
