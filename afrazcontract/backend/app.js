const grpc = require('@grpc/grpc-js');
const { connect, hash, signers } = require('@hyperledger/fabric-gateway');
const crypto = require('node:crypto');
const fs = require('node:fs/promises');
const path = require('node:path');
const { TextDecoder } = require('node:util');
const express = require('express');

const app = express();
const port =  3000;

app.use(express.json());

// Environment variables
const channelName = process.env.CHANNEL_NAME || 'mychannel';
const chaincodeName = process.env.CHAINCODE_NAME || 'mycc';
const mspId = process.env.MSP_ID || 'Org1MSP';

// Helper function to get environment variables with defaults
function envOrDefault(key, defaultValue) {
    return process.env[key] || defaultValue;
}

const cryptoPath = envOrDefault(
    'CRYPTO_PATH',
    path.resolve(
        __dirname,
        '..',
        '..',
        '..',
        'test-network',
        'organizations',
        'peerOrganizations',
        'org1.example.com'
    )
);

// Path to user private key directory.
const keyDirectoryPath = envOrDefault(
    'KEY_DIRECTORY_PATH',
    path.resolve(
        cryptoPath,
        'users',
        'User1@org1.example.com',
        'msp',
        'keystore'
    )
);

// Path to user certificate directory.
const certDirectoryPath = envOrDefault(
    'CERT_DIRECTORY_PATH',
    path.resolve(
        cryptoPath,
        'users',
        'User1@org1.example.com',
        'msp',
        'signcerts'
    )
);

// Path to peer tls certificate.
const tlsCertPath = process.env.TLS_CERT_PATH || path.resolve(
    __dirname, '..', '..', '..', 'test-network', 'organizations', 'peerOrganizations', 'org1.example.com', 'peers', 'peer0.org1.example.com', 'tls', 'ca.crt'
);

const peerEndpoint = process.env.PEER_ENDPOINT || 'localhost:7051';
const peerHostAlias = process.env.PEER_HOST_ALIAS || 'peer0.org1.example.com';

const utf8Decoder = new TextDecoder();
let contract;

(async () => {
    // Display connection parameters
    console.log(`Channel: ${channelName}`);
    console.log(`Chaincode: ${chaincodeName}`);
    console.log(`MSP ID: ${mspId}`);
    console.log(`Peer endpoint: ${peerEndpoint}`);

    // Create gRPC connection
    const client = await newGrpcConnection();
    const gateway = connect({
        client,
        identity: await newIdentity(),
        signer: await newSigner(),
        hash: hash.sha256,
        evaluateOptions: () => ({ deadline: Date.now() + 5000 }),
        endorseOptions: () => ({ deadline: Date.now() + 15000 }),
        submitOptions: () => ({ deadline: Date.now() + 5000 }),
        commitStatusOptions: () => ({ deadline: Date.now() + 60000 }),
    });

    const network = gateway.getNetwork(channelName);
    contract = network.getContract(chaincodeName);
    console.log('Connected to Fabric network');
})();

// Routes
app.post('/initLedger', async (req, res) => {
    try {
        await contract.submitTransaction('InitLedger');
        res.json({ message: 'Ledger initialized with sample identity' });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

app.post('/identities', async (req, res) => {
    try {
        const { id, title, firstName, lastName, cnic, dob, gender, mobile } = req.body;
        
        if (!id || !firstName || !lastName || !cnic || !dob || !gender || !mobile) {
            return res.status(400).json({ error: 'Missing required fields' });
        }

        await contract.submitTransaction(
            'CreateIdentity',
            id,
            title || '',
            firstName,
            '',
            lastName,
            cnic,
            dob,
            gender,
            mobile
        );
        
        res.json({ message: 'Identity created successfully', id });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

app.get('/identities/:id', async (req, res) => {
    try {
        const id = req.params.id;
        const resultBytes = await contract.evaluateTransaction('ReadIdentity', id);
        const resultJson = utf8Decoder.decode(resultBytes);
        const result = JSON.parse(resultJson);
        res.json(result);
    } catch (error) {
        res.status(404).json({ error: error.message });
    }
});

app.put('/identities/:id', async (req, res) => {
    try {
        const id = req.params.id;
        const { mobile, address } = req.body;
        
        if (!mobile && !address) {
            return res.status(400).json({ error: 'No fields to update' });
        }

        await contract.submitTransaction(
            'UpdateIdentity',
            id,
            mobile || '',
            address || ''
        );
        
        res.json({ message: 'Identity updated successfully' });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

app.get('/identities', async (req, res) => {
    try {
        const resultBytes = await contract.evaluateTransaction('GetAllIdentities');
        const resultJson = utf8Decoder.decode(resultBytes);
        const result = JSON.parse(resultJson);
        res.json(result);
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

app.delete('/identities/:id', async (req, res) => {
    try {
        const id = req.params.id;
        await contract.submitTransaction('DeleteIdentity', id);
        res.json({ message: 'Identity deleted successfully' });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

// Start server
app.listen(port, () => {
    console.log(`API server running on port ${port}`);
});

// Helper functions
async function newGrpcConnection() {
    const tlsRootCert = await fs.readFile(tlsCertPath);
    const tlsCredentials = grpc.credentials.createSsl(tlsRootCert);
    return new grpc.Client(peerEndpoint, tlsCredentials, {
        'grpc.ssl_target_name_override': peerHostAlias,
    });
}

async function newIdentity() {
    const certPath = await getFirstDirFileName(certDirectoryPath);
    const credentials = await fs.readFile(certPath);
    return { mspId, credentials };
}

async function getFirstDirFileName(dirPath) {
    const files = await fs.readdir(dirPath);
    const file = files[0];
    if (!file) {
        throw new Error(`No files in directory: ${dirPath}`);
    }
    return path.join(dirPath, file);
}

async function newSigner() {
    const keyPath = await getFirstDirFileName(keyDirectoryPath);
    const privateKeyPem = await fs.readFile(keyPath);
    const privateKey = crypto.createPrivateKey(privateKeyPem);
    return signers.newPrivateKeySigner(privateKey);
}