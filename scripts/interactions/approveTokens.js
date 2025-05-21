const { task } = require("hardhat/config");
const fs = require('fs');
const path = require('path');

task("approve-tokens", "Approves a spender to withdraw BRLT tokens from the signer's account")
  .addParam("proxy", "Proxy address of the BRLT contract")
  .addParam("spender", "Address of the account to approve")
  .addParam("amount", "Amount of BRLT to approve (e.g., 1000 for 1000 BRLT)")
  .setAction(async (taskArgs, hre) => {
    const { ethers, network } = hre;
    const [ownerSigner] = await ethers.getSigners(); // The owner of the tokens, approving the spender

    const { proxy: proxyAddress, spender, amount: amountToApproveStr } = taskArgs;

    if (!ethers.isAddress(proxyAddress)) {
      console.error(`Error: Invalid proxy address provided: ${proxyAddress}`);
      throw new Error("Invalid proxy address");
    }
    if (!ethers.isAddress(spender)) {
      console.error(`Error: Invalid spender address provided: ${spender}`);
      throw new Error("Invalid spender address");
    }

    let amountInWei;
    try {
        amountInWei = ethers.parseUnits(amountToApproveStr, 18); // BRLT has 18 decimals
    } catch (e) {
        console.error(`Error: Invalid amount specified: ${amountToApproveStr}. Must be a number.`);
        throw e;
    }

    console.log(`Executing approve-tokens task for BRLT at ${proxyAddress} on network ${network.name}`);
    console.log(`Owner (signer): ${ownerSigner.address}`);
    console.log(`Spender: ${spender}`);
    console.log(`Amount to approve: ${amountToApproveStr} BRLT (${amountInWei.toString()} wei)\n`);

    let brlt;
    try {
        const abiPath = path.join(__dirname, "../../artifacts/solidity/BRLT.sol/BRLT.json");
        const brltArtifact = JSON.parse(fs.readFileSync(abiPath, 'utf8'));
        brlt = await ethers.getContractAt(brltArtifact.abi, proxyAddress, ownerSigner);
    } catch (e) {
        console.error(`Error attaching to BRLT contract at ${proxyAddress}:`, e);
        throw e;
    }

    try {
        console.log(`Current allowance for ${spender} from ${ownerSigner.address}: ${ethers.formatUnits(await brlt.allowance(ownerSigner.address, spender), 18)} BRLT`);
        
        console.log(`Attempting to approve ${spender} for ${amountToApproveStr} BRLT from ${ownerSigner.address}...`);
        const tx = await brlt.approve(spender, amountInWei);
        console.log(`Transaction sent: ${tx.hash}`);
        await tx.wait();
        console.log("Transaction confirmed.");

        const newAllowance = await brlt.allowance(ownerSigner.address, spender);
        console.log(`New allowance for ${spender} from ${ownerSigner.address}: ${ethers.formatUnits(newAllowance, 18)} BRLT`);
        console.log(`Successfully approved ${spender} for ${amountToApproveStr} BRLT.`);

    } catch (error) {
        console.error(`Error during approval:`, error);
        throw error;
    }
  }); 