# loadshedder

 This library is used to shed excess load on the basis of cpu utilisation. You can use instance of prometheus GaugeVec & CounterVec if you want to plot CPU utilisation & sheddings metric (they are optional - you can use nil if you don't want to add signal for them). To get the library simply run : 

    go get -u github.com/anandshukla-sharechat/loadshedder@main

## NewSheddingStat : 
    use this function to create an instance of shedding stat. Names for the cpu stat gauge vec & load shedding counter vec

    totalCpuUsageStat                 = "total_cpu_usage_stat"
	loadSheddingMetric                = "Load_shedding_metrics"

 

## GinUnarySheddingInterceptor : 
    use this function as middleware. It needs sheddingStat, cpu threshold and probe API (for disabling load shedding for live & readiness check api endpoint)


##    Example :
### In main.go add these lines 


    sheddingStat := loadshedder.NewSheddingStat(loadSheddingMetrics *prometheus.CounterVec, cpuMetrics *prometheus.GaugeVec) // returns object of *SheddingStat type
    router.Use(loadshedder.GinUnarySheddingInterceptor(shedderEnabled bool, cpuThreshold int64, probeAPI string, sheddingStat *SheddingStat))

    
shedderEnabled : boolean flag to enable load shedder or not

cpuThreshold : the threshold beyond which load shedding starts

probeAPI : To disable load shedding for api which is responsible for liveness/readiness probe checks 


## Queries for Prometheus :

Query for load shedding stats : 

    100 * (sum(rate(Load_shedding_metrics{service_prometheus_track=~"livestream-feed-relevance", stat = "dropped"}[2m])) by (kubernetes_pod_name) OR on() vector(0)) / (sum(rate(Load_shedding_metrics{service_prometheus_track=~"livestream-feed-relevance", stat = "total"}[2m])) by (kubernetes_pod_name) OR on() vector(0))

Query for CPU utilisation :

    avg(total_cpu_usage_stat{service_prometheus_track=~"livestream-feed-relevance", metric = "usage"}[2m]) by (kubernetes_pod_name) 
    
