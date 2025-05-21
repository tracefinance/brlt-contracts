const { task } = require("hardhat/config");
const fs = require('fs');
const path = require('path');

task("mint-tokens", "Mints BRLT tokens to a specified recipient")
  .addParam("proxy", "Proxy address of the BRLT contract")
  .addParam("recipient", "Address of the recipient")
  .addParam("amount", "Amount of BRLT to mint (e.g., 1000 for 1000 BRLT)")
  .setAction(async (taskArgs, hre) => {
    const { ethers, network } = hre;
    const [deployer] = await ethers.getSigners(); // Signer must have MINTER_ROLE

    const { proxy: proxyAddress, recipient, amount: amountToMintStr } = taskArgs;

    if (!ethers.isAddress(proxyAddress)) {
      console.error(`Error: Invalid proxy address provided: ${proxyAddress}`);
      throw new Error("Invalid proxy address");
    }
    if (!ethers.isAddress(recipient)) {
      console.error(`Error: Invalid recipient address provided: ${recipient}`);
      throw new Error("Invalid recipient address");
    }

    let amountInWei;
    try {
        amountInWei = ethers.parseUnits(amountToMintStr, 18); // BRLT has 18 decimals
    } catch (e) {
        console.error(`Error: Invalid amount specified: ${amountToMintStr}. Must be a number.`);
        throw e;
    }

    console.log(`Executing mint-tokens task for BRLT at ${proxyAddress} on network ${network.name}`);
    console.log(`Recipient: ${recipient}`);
    console.log(`Amount to mint: ${amountToMintStr} BRLT (${amountInWei.toString()} wei)`);
    console.log(`Signer (potential minter): ${deployer.address}\n`);

    let brlt;
    try {
        const abiPath = path.join(__dirname, "../../artifacts/solidity/BRLT.sol/BRLT.json");
        const brltArtifact = JSON.parse(fs.readFileSync(abiPath, 'utf8'));
        brlt = await ethers.getContractAt(brltArtifact.abi, proxyAddress, deployer);
    } catch (e) {
        console.error(`Error attaching to BRLT contract at ${proxyAddress}:`, e);
        throw e;
    }

    try {
        console.log(`Attempting to mint ${amountToMintStr} BRLT to ${recipient}...`);
        
        const minterRole = await brlt.MINTER_ROLE();
        const hasMinterRole = await brlt.hasRole(minterRole, deployer.address);
        if (!hasMinterRole) {
            console.error(`Error: Signer ${deployer.address} does not have the MINTER_ROLE.`);
            console.error(`Please grant MINTER_ROLE to this address using the manage-roles task before minting.`);
            throw new Error(`Signer ${deployer.address} does not have MINTER_ROLE`);
        }

        const balanceBefore = await brlt.balanceOf(recipient);
        console.log(`Balance of ${recipient} before mint: ${ethers.formatUnits(balanceBefore, 18)} BRLT`);

        const tx = await brlt.mint(recipient, amountInWei);
        console.log(`Transaction sent: ${tx.hash}`);
        await tx.wait();
        console.log("Transaction confirmed.");

        const balanceAfter = await brlt.balanceOf(recipient);
        console.log(`Balance of ${recipient} after mint: ${ethers.formatUnits(balanceAfter, 18)} BRLT`);
        console.log(`Successfully minted ${amountToMintStr} BRLT to ${recipient}.`);

    } catch (error) {
        console.error(`Error during minting:`, error);
        throw error;
    }
  }); 