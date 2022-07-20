FROM quay.io/stackstate/trace-java-demo:books-app-master

COPY ./pom.xml /app/pom.xml
RUN cd /app && mvn dependency:resolve && cd ..

ENTRYPOINT ["/entrypoint.sh"]
