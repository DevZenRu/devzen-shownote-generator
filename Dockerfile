# syntax=docker/dockerfile:1

FROM sbtscala/scala-sbt:eclipse-temurin-11.0.16_1.8.0_2.12.17 AS builder
ADD . /app/
RUN cd /app && sbt compile stage

FROM eclipse-temurin:11 AS app
RUN mkdir -p /shownotegen
COPY --from=builder /app/target/universal/stage /shownotegen/
WORKDIR /shownotegen
ENV JAVA_OPTS="-Xms128m -XX:+UseContainerSupport -XX:MaxRAMPercentage=60.0 -XX:+UseContainerSupport"
ENV JVM_OPTS="-Xms128m -XX:+UseContainerSupport -XX:MaxRAMPercentage=60.0 -XX:+UseContainerSupport"
ENTRYPOINT bin/devzen-shownote-generator
