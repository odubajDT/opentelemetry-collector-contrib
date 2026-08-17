FROM mcr.microsoft.com/mssql/server:2025-latest@sha256:4bab24f36c1ecd48e85f7d37df26e6bf301641d84c3fe652f9a0dcc947d512e1

ENV ACCEPT_EULA=Y
ENV SA_PASSWORD=YourStrong!Passw0rd

# Set the working directory
WORKDIR /usr/src/app

# Copy the scripts and set the executable permissions before copying them into the image
COPY entrypoint.sh configure-db.sh init.sql ./

ENTRYPOINT ["/bin/bash", "/usr/src/app/entrypoint.sh"]