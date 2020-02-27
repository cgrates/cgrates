FROM scratch
ADD ./cgr-engine /
ADD ./cgrates.json /etc/cgrates/
CMD ["/cgr-engine","--logger=*stdout","-httprof_path=/pprof"]
