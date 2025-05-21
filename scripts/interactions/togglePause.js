const { task } = require("hardhat/config");
const fs = require('fs');
const path = require('path');

task("toggle-pause", "Pauses or unpauses the BRLT contract")
  .addParam("proxy", "Proxy address of the BRLT contract")
  .addParam("action", "The action to perform: 'pause' or 'unpause'")
  .setAction(async (taskArgs, hre) => {
    const { ethers, network } = hre;
    const [pauserSigner] = await ethers.getSigners(); // Signer must have PAUSER_ROLE

    const { proxy: proxyAddress, action } = taskArgs;

    if (action !== 'pause' && action !== 'unpause') {
      console.error("Error: Action must be 'pause' or 'unpause'.");
      throw new Error("Invalid action parameter");
    }
    if (!ethers.isAddress(proxyAddress)) {
      console.error(`Error: Invalid proxy address provided: ${proxyAddress}`);
      throw new Error("Invalid proxy address");
    }

    console.log(`Executing toggle-pause task for BRLT at ${proxyAddress} on network ${network.name}`);
    console.log(`Action: ${action}`);
    console.log(`Signer (potential pauser): ${pauserSigner.address}\n`);

    let brlt;
    try {
        const abiPath = path.join(__dirname, "../../artifacts/solidity/BRLT.sol/BRLT.json");
        const brltArtifact = JSON.parse(fs.readFileSync(abiPath, 'utf8'));
        brlt = await ethers.getContractAt(brltArtifact.abi, proxyAddress, pauserSigner);
    } catch (e) {
        console.error(`Error attaching to BRLT contract at ${proxyAddress}:`, e);
        throw e;
    }

    try {
        const pauserRole = await brlt.PAUSER_ROLE();
        const hasPauserRole = await brlt.hasRole(pauserRole, pauserSigner.address);
        if (!hasPauserRole) {
            console.error(`Error: Signer ${pauserSigner.address} does not have the PAUSER_ROLE.`);
            console.error(`Please grant PAUSER_ROLE to this address using the manage-roles task.`);
            throw new Error(`Signer ${pauserSigner.address} does not have PAUSER_ROLE`);
        }

        const currentPausedState = await brlt.paused();
        console.log(`Contract paused state before action: ${currentPausedState}`);

        let tx;
        if (action === 'pause') {
            if (currentPausedState) {
                console.log("Contract is already paused. No action needed.");
                return;
            }
            console.log("Pausing contract...");
            tx = await brlt.pause();
        } else { // action === 'unpause'
            if (!currentPausedState) {
                console.log("Contract is already unpaused. No action needed.");
                return;
            }
            console.log("Unpausing contract...");
            tx = await brlt.unpause();
        }

        console.log(`Transaction sent: ${tx.hash}`);
        await tx.wait();
        console.log("Transaction confirmed.");

        const newPausedState = await brlt.paused();
        console.log(`Contract paused state after action: ${newPausedState}`);
        console.log(`Successfully performed action: ${action}.`);

    } catch (error) {
        console.error(`Error during toggle pause action:`, error);
        throw error;
    }
  }); 