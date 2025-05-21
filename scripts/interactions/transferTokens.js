const { task } = require("hardhat/config");
const fs = require('fs');
const path = require('path');

task("transfer-tokens", "Transfers BRLT tokens from the signer's account to a recipient")
  .addParam("proxy", "Proxy address of the BRLT contract")
  .addParam("to", "Address of the recipient")
  .addParam("amount", "Amount of BRLT to transfer (e.g., 100 for 100 BRLT)")
  .setAction(async (taskArgs, hre) => {
    const { ethers, network } = hre;
    const [senderSigner] = await ethers.getSigners(); // The owner/sender of the tokens

    const { proxy: proxyAddress, to: recipientAddress, amount: amountToTransferStr } = taskArgs;

    if (!ethers.isAddress(proxyAddress)) {
      console.error(`Error: Invalid proxy address provided: ${proxyAddress}`);
      throw new Error("Invalid proxy address");
    }
    if (!ethers.isAddress(recipientAddress)) {
      console.error(`Error: Invalid recipient address provided: ${recipientAddress}`);
      throw new Error("Invalid recipient address");
    }

    let amountInWei;
    try {
        amountInWei = ethers.parseUnits(amountToTransferStr, 18); // BRLT has 18 decimals
    } catch (e) {
        console.error(`Error: Invalid amount specified: ${amountToTransferStr}. Must be a number.`);
        throw e;
    }

    console.log(`Executing transfer-tokens task for BRLT at ${proxyAddress} on network ${network.name}`);
    console.log(`Sender (signer): ${senderSigner.address}`);
    console.log(`Recipient: ${recipientAddress}`);
    console.log(`Amount to transfer: ${amountToTransferStr} BRLT (${amountInWei.toString()} wei)\n`);

    let brlt;
    try {
        const abiPath = path.join(__dirname, "../../artifacts/solidity/BRLT.sol/BRLT.json");
        const brltArtifact = JSON.parse(fs.readFileSync(abiPath, 'utf8'));
        brlt = await ethers.getContractAt(brltArtifact.abi, proxyAddress, senderSigner);
    } catch (e) {
        console.error(`Error attaching to BRLT contract at ${proxyAddress}:`, e);
        throw e;
    }

    try {
        const senderBalanceBefore = await brlt.balanceOf(senderSigner.address);
        console.log(`Sender balance before transfer: ${ethers.formatUnits(senderBalanceBefore, 18)} BRLT`);
        const recipientBalanceBefore = await brlt.balanceOf(recipientAddress);
        console.log(`Recipient balance before transfer: ${ethers.formatUnits(recipientBalanceBefore, 18)} BRLT`);

        if (senderBalanceBefore < amountInWei) {
            console.error(`Error: Sender ${senderSigner.address} has insufficient balance. Has ${ethers.formatUnits(senderBalanceBefore, 18)}, needs ${amountToTransferStr}.`);
            throw new Error("Insufficient sender balance");
        }
        
        console.log(`Attempting to transfer ${amountToTransferStr} BRLT from ${senderSigner.address} to ${recipientAddress}...`);
        const tx = await brlt.transfer(recipientAddress, amountInWei);
        console.log(`Transaction sent: ${tx.hash}`);
        await tx.wait();
        console.log("Transaction confirmed.");

        const senderBalanceAfter = await brlt.balanceOf(senderSigner.address);
        console.log(`Sender balance after transfer: ${ethers.formatUnits(senderBalanceAfter, 18)} BRLT`);
        const recipientBalanceAfter = await brlt.balanceOf(recipientAddress);
        console.log(`Recipient balance after transfer: ${ethers.formatUnits(recipientBalanceAfter, 18)} BRLT`);
        console.log(`Successfully transferred ${amountToTransferStr} BRLT to ${recipientAddress}.`);

    } catch (error) {
        console.error(`Error during transfer:`, error);
        throw error;
    }
  }); 