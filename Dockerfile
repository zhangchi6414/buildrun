FROM harbor.sinodata.vip/chaincode-external/alpine:v1.0.0
COPY main /root/
WORKDIR /root
CMD ./buildRun
