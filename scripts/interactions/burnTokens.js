const { task } = require("hardhat/config");
const fs = require('fs');
const path = require('path');

task("burn-tokens", "Burns BRLT tokens from a specified account using allowance (burnFrom)")
  .addParam("proxy", "Proxy address of the BRLT contract")
  .addParam("from", "Address of the account to burn tokens from")
  .addParam("amount", "Amount of BRLT to burn (e.g., 500 for 500 BRLT)")
  .setAction(async (taskArgs, hre) => {
    const { ethers, network } = hre;
    const [burnerSigner] = await ethers.getSigners(); // Signer must have BURNER_ROLE and allowance

    const { proxy: proxyAddress, from: burnFromAccount, amount: amountToBurnStr } = taskArgs;

    if (!ethers.isAddress(proxyAddress)) {
      console.error(`Error: Invalid proxy address provided: ${proxyAddress}`);
      throw new Error("Invalid proxy address");
    }
    if (!ethers.isAddress(burnFromAccount)) {
      console.error(`Error: Invalid 'from' address provided: ${burnFromAccount}`);
      throw new Error("Invalid 'from' address");
    }

    let amountInWei;
    try {
        amountInWei = ethers.parseUnits(amountToBurnStr, 18); // BRLT has 18 decimals
    } catch (e) {
        console.error(`Error: Invalid amount specified: ${amountToBurnStr}. Must be a number.`);
        throw e;
    }

    console.log(`Executing burn-tokens task for BRLT at ${proxyAddress} on network ${network.name}`);
    console.log(`Account to burn from: ${burnFromAccount}`);
    console.log(`Amount to burn: ${amountToBurnStr} BRLT (${amountInWei.toString()} wei)`);
    console.log(`Signer (potential burner): ${burnerSigner.address}\n`);

    let brlt;
    try {
        const abiPath = path.join(__dirname, "../../artifacts/solidity/BRLT.sol/BRLT.json");
        const brltArtifact = JSON.parse(fs.readFileSync(abiPath, 'utf8'));
        brlt = await ethers.getContractAt(brltArtifact.abi, proxyAddress, burnerSigner);
    } catch (e) {
        console.error(`Error attaching to BRLT contract at ${proxyAddress}:`, e);
        throw e;
    }

    try {
        const burnerRole = await brlt.BURNER_ROLE();
        const hasBurnerRole = await brlt.hasRole(burnerRole, burnerSigner.address);
        if (!hasBurnerRole) {
            console.error(`Error: Signer ${burnerSigner.address} does not have the BURNER_ROLE.`);
            console.error(`Please grant BURNER_ROLE to this address using the manage-roles task before burning.`);
            throw new Error(`Signer ${burnerSigner.address} does not have BURNER_ROLE`);
        }

        // Check allowance
        const allowance = await brlt.allowance(burnFromAccount, burnerSigner.address);
        console.log(`Current allowance for ${burnerSigner.address} to spend from ${burnFromAccount}: ${ethers.formatUnits(allowance, 18)} BRLT`);

        if (allowance < amountInWei) {
            console.error(`Error: Insufficient allowance. ${burnFromAccount} needs to approve ${burnerSigner.address} for at least ${amountToBurnStr} BRLT.`);
            console.error(`You can do this via Hardhat console or another wallet. Example for Hardhat console (ensure correct signer for 'burnFromAccount'):`);
            console.error(`  // const BRLT = await hre.ethers.getContractFactory("BRLT");`);
            console.error(`  // const brlt = BRLT.attach("${proxyAddress}");`);
            console.error(`  // const targetSigner = await hre.ethers.getSigner("${burnFromAccount}"); // May need specific setup if not one of the default hardhat accounts`);
            console.error(`  // await brlt.connect(targetSigner).approve("${burnerSigner.address}", hre.ethers.parseUnits("${amountToBurnStr}", 18));`);
            throw new Error("Insufficient allowance");
        }

        console.log(`Attempting to burn ${amountToBurnStr} BRLT from ${burnFromAccount}...`);
        const balanceBefore = await brlt.balanceOf(burnFromAccount);
        console.log(`Balance of ${burnFromAccount} before burn: ${ethers.formatUnits(balanceBefore, 18)} BRLT`);
        const totalSupplyBefore = await brlt.totalSupply();
        console.log(`Total supply before burn: ${ethers.formatUnits(totalSupplyBefore, 18)} BRLT`);

        const tx = await brlt.burnFrom(burnFromAccount, amountInWei);
        console.log(`Transaction sent: ${tx.hash}`);
        await tx.wait();
        console.log("Transaction confirmed.");

        const balanceAfter = await brlt.balanceOf(burnFromAccount);
        console.log(`Balance of ${burnFromAccount} after burn: ${ethers.formatUnits(balanceAfter, 18)} BRLT`);
        const totalSupplyAfter = await brlt.totalSupply();
        console.log(`Total supply after burn: ${ethers.formatUnits(totalSupplyAfter, 18)} BRLT`);
        console.log(`Successfully burned ${amountToBurnStr} BRLT from ${burnFromAccount}.`);

    } catch (error) {
        console.error(`Error during burning:`, error);
        throw error;
    }
  }); 