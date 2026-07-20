FROM alpine:3.21 AS provider
RUN apk add --no-cache curl
ARG EMAIL_OTP_JAR_URL=https://github.com/mesutpiskin/keycloak-2fa-email-authenticator/releases/download/v26.4.3/keycloak-2fa-email-authenticator-v26.4.3-KC26.6.3.jar
RUN curl -fsSL -o /tmp/keycloak-2fa-email-authenticator.jar "${EMAIL_OTP_JAR_URL}" \
  && unzip -t /tmp/keycloak-2fa-email-authenticator.jar >/dev/null

FROM quay.io/keycloak/keycloak:26.6.3
COPY --from=provider /tmp/keycloak-2fa-email-authenticator.jar \
  /opt/keycloak/providers/keycloak-2fa-email-authenticator.jar
