FROM eclipse-temurin:21-jdk AS build
WORKDIR /build
COPY Main.java .
RUN javac Main.java -d /build/classes

FROM gcr.io/distroless/java21-debian12:nonroot
WORKDIR /app
COPY --from=build /build/classes /app/classes

EXPOSE 8080
ENTRYPOINT ["java", "-cp", "/app/classes", "Main"]
