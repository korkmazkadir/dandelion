FROM algorand/testnet

# installs unzip && cleans up apt cache && Removes default data folder
RUN apt-get update && \
    apt-get install -y unzip iproute2 jq python3  && rm -rf /var/lib/apt/lists/* && rm -r /root/node/data/


WORKDIR /root/node

# Adds node folder to path to access algorand binaries
ENV PATH="/root/node:${PATH}"


# Dowloads dandelion and unzip it
#RUN curl --location --remote-header-name --remote-name https://github.com/korkmazkadir/dandelion/releases/download/v0.0.1-beta/dandelion_linux-amd64.zip \
#&& unzip dandelion_linux-amd64.zip \
#&& rm dandelion_linux-amd64.zip

RUN curl -sSL https://raw.githubusercontent.com/thombashi/tcconfig/master/scripts/installer.sh | bash


COPY  dandelion algod /root/node/

#ENV MAX_TXN_BYTES_PER_BLOCK 10000000
#ENV MAX_TXN_NOTE_BYTES 10240


ENTRYPOINT ["dandelion", "run-node"]
