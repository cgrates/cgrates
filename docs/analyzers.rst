# AnalyzerS  


**AnalyzerS** is the service component within **CGRateS** responsible for capturing and indexing API calls, enabling subsequent querying and analysis of API interactions.


```
"analyzers":{                                   
    "enabled": false,                           
    "db_path": "/var/spool/cgrates/analyzers",  // path to the folder where to store the information
    "index_type": "*scorch",                    // the type of index used for storage: 
    "ttl": "24h",                               // duration until captures expire                  
    "cleanup_interval": "1h",                   // interval at which expired captures are removed
}
```
The settings above represent the defaults. To use AnalyzerS the "enabled" field should be set to true.

# How to use AnalyzerS
Let's assume that we have previously called the following APIs:

```
{"method":"CoreSv1.Status","params":[{"Tenant":"cgrates.org"}],"id":1}
{"method":"APIerSv1.SetTPTiming","params":[{"TPid":"balance","ID":"monthly","Years":"*any","Months":"*any","MonthDays":"1","WeekDays":"*any","Time":"00:00:00"}],"id":2}
{"method":"APIerSv1.SetTPTiming","params":[{"TPid":"balance","ID":"monthly","Years":"*any","Months":"*any","MonthDays":"30","WeekDays":"*any","Time":"00:00:00"}],"id":3}
```

If AnalyzerS is enabled, the CoreSv1.Status API the 2 APIerSv1.SetTPTiming APIs have been captured. For including all APIs in our search we use AnalyzerSv1.StringQuery method with empty parameters, and it has the following structure:

```
{                       
    "method": "AnalyzerSv1.StringQuery",
    "params": [{}],
    "id": 4
}
```
and we would receive the following as a reply:

```
{
    "id": 4,
    "result": [{
        "Reply": {
            "active_memory": "8.1MiB",
            "cpu_time": "1.69s",
            "go_version": "go1.23.2",
            "goroutines": 58,
            "node_id": "e0a851b",
            "open_files": 26,
            "os_threads_in_use": 8,
            "pid": 788,
            "resident_memory": "52.4MiB",
            "running_since": "Wed Jan 22 13:16:58 UTC 2025",
            "system_memory": "21.3MiB",
            "version": "CGRateS@v0.11.0~dev-20250121183033-c171937c3d8e"
        },
        "ReplyError": null,
        "RequestDestination": "[::1]:2012",
        "RequestDuration": "10.688536ms",
        "RequestEncoding": "*json",
        "RequestID": 1,
        "RequestMethod": "CoreSv1.Status",
        "RequestParams": {
            "Debug": false,
            "Timezone": "",
            "Tenant": "cgrates.org",
            "APIOpts": null
        },
        "RequestSource": "[::1]:41370",
        "RequestStartTime": "2025-01-22T13:23:53Z"
    }, {
        "Reply": "OK",
        "ReplyError": null,
        "RequestDestination": "[::1]:2012",
        "RequestDuration": "25.427147ms",
        "RequestEncoding": "*json",
        "RequestID": 2,
        "RequestMethod": "APIerSv1.SetTPTiming",
        "RequestParams": {
            "TPid": "balance",
            "ID": "monthly",
            "Years": "*any",
            "Months": "*any",
            "MonthDays": "1",
            "WeekDays": "*any",
            "Time": "00:00:00"
        },
        "RequestSource": "[::1]:41370",
        "RequestStartTime": "2025-01-22T13:24:55Z"
    }, {
        "Reply": "OK",
        "ReplyError": null,
        "RequestDestination": "[::1]:2012",
        "RequestDuration": "17.401841ms",
        "RequestEncoding": "*json",
        "RequestID": 3,
        "RequestMethod": "APIerSv1.SetTPTiming",
        "RequestParams": {
            "TPid": "balance",
            "ID": "monthly",
            "Years": "*any",
            "Months": "*any",
            "MonthDays": "30",
            "WeekDays": "*any",
            "Time": "00:00:00"
        },
        "RequestSource": "[::1]:41370",
        "RequestStartTime": "2025-01-22T13:25:49Z"
    }],
    "error": null
}
```
But if we want to query only for APIerSv1.SetTPTiming APIs we add HeaderFilter as a parameter, and now the AnalyzerSv1.StringQuery method has this structure:

```
{
    "method": "AnalyzerSv1.StringQuery",
    "params": [{
        "HeaderFilters": "+RequestMethod:\"APIerSv1.SetTPTiming\""
    }],
    "id": 4
}
```
we would receive the following as a reply:

```
[{
    "id": 5,
    "result": [{
        "Reply": "OK",
        "ReplyError": null,
        "RequestDestination": "[::1]:2012",
        "RequestDuration": "25.427147ms",
        "RequestEncoding": "*json",
        "RequestID": 2,
        "RequestMethod": "APIerSv1.SetTPTiming",
        "RequestParams": {
            "TPid": "balance",
            "ID": "monthly",
            "Years": "*any",
            "Months": "*any",
            "MonthDays": "1",
            "WeekDays": "*any",
            "Time": "00:00:00"
        },
        "RequestSource": "[::1]:41370",
        "RequestStartTime": "2025-01-22T13:24:55Z"
    }, {
        "Reply": "OK",
        "ReplyError": null,
        "RequestDestination": "[::1]:2012",
        "RequestDuration": "17.401841ms",
        "RequestEncoding": "*json",
        "RequestID": 3,
        "RequestMethod": "APIerSv1.SetTPTiming",
        "RequestParams": {
            "TPid": "balance",
            "ID": "monthly",
            "Years": "*any",
            "Months": "*any",
            "MonthDays": "30",
            "WeekDays": "*any",
            "Time": "00:00:00"
        },
        "RequestSource": "[::1]:41370",
        "RequestStartTime": "2025-01-22T13:25:49Z"
    }],
    "error": null
}]
```
Looking at the reply, it's easier to see the fields that filtering can be based on when using HeaderFilters: RequestDuration, RequestEncoding, RequestMethod etc. Their structure is dictated by [bleve's query language](http://blevesearch.com/docs/Query-String-Query/).
Using this type of filter is recommended for improved querying performance.

If we also need to filter the contents on the APIerSv1.SetTPTiming APIs then we need to add another parameter in AnalyzerSv1.StringQuery method which is ContentFilter, adding that parameter now we have the following structure:

```
{
    "method": "AnalyzerSv1.StringQuery",
    "params": [{
        "HeaderFilters": "+RequestMethod:\"APIerSv1.SetTPTiming\"",
        "ContentFilters": [
            "*string:~*req.MonthDays:30"
        ]
    }],
    "id": 6
}
```
and we will receive the following as a reply:

```
{
    "id": 6,
    "result": [{
        "Reply": "OK",
        "ReplyError": null,
        "RequestDestination": "[::1]:2012",
        "RequestDuration": "17.401841ms",
        "RequestEncoding": "*json",
        "RequestID": 3,
        "RequestMethod": "APIerSv1.SetTPTiming",
        "RequestParams": {
            "TPid": "balance",
            "ID": "monthly",
            "Years": "*any",
            "Months": "*any",
            "MonthDays": "30",
            "WeekDays": "*any",
            "Time": "00:00:00"
        },
        "RequestSource": "[::1]:41370",
        "RequestStartTime": "2025-01-22T13:25:49Z"
    }],
    "error": null
}
```


 **ContentFilters** have the structure of normal CGRateS inline filters and can be of 4 types:
- `*req` - filtering through request fields
- `*opts`- filtering through APIOpts fields
- `*rep` - filtering through reply fields
- `*hdr` - filtering through header fields
ContentFilters are used for filtering the contents of requests/replies/headers/opts. These are slower compared to HeaderFilters.