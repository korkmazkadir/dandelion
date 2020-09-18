FROM algorand/testnet

# installs unzip
RUN apt-get update && \
    apt-get install -y unzip

WORKDIR /root/node


# Adds node folder to path to access algorand binaries
ENV PATH="/root/node:${PATH}"

# Removes data folder
RUN rm -r ./data/


# Dowloads dandelion and unzip it
RUN curl --location --remote-header-name --remote-name https://github.com/korkmazkadir/dandelion/releases/download/v1.0.0-beta/dandelion_linux-amd64.zip \
&& unzip dandelion_linux-amd64.zip \
&& rm dandelion_linux-amd64.zip


ENTRYPOINT ["./dandelion", "run-node"]

