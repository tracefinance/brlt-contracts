const { task } = require("hardhat/config");

task("deploy-contract", "Deploys the BRLT contract")
  .addParam("contractName", "The name of the contract to deploy (Currently only BRLT is supported)")
  .setAction(async (taskArgs, hre) => {
    const { contractName } = taskArgs;
    console.log(`Attempting to deploy contract: ${contractName}`);

    if (contractName === "BRLT") {
      console.log("Deploying BRLT...");
      const [deployer] = await hre.ethers.getSigners();
      const deployerAddress = await deployer.getAddress();

      let initialAdmin = process.env.BRLT_INITIAL_ADMIN;
      if (!initialAdmin || initialAdmin.trim() === "") {
        console.log("BRLT_INITIAL_ADMIN not set in .env, using deployer address as initial admin.");
        initialAdmin = deployerAddress;
      } else {
        initialAdmin = initialAdmin.trim();
        console.log(`Using BRLT_INITIAL_ADMIN from .env: ${initialAdmin}`);
      }

      if (!hre.ethers.isAddress(initialAdmin)) {
        throw new Error(`Invalid BRLT_INITIAL_ADMIN address: ${initialAdmin}`);
      }

      const BRLT = await hre.ethers.getContractFactory("BRLT");

      console.log("Network:", hre.network.name);
      console.log("Deployer Address:", deployerAddress);
      console.log("Initial Admin for BRLT:", initialAdmin);
      console.log("Deploying BRLT as UUPS Proxy...");

      const brlt = await hre.upgrades.deployProxy(BRLT, [initialAdmin], {
          kind: 'uups',
          initializer: 'initialize',
          timeout: 600000, // 600 seconds = 10 minutes
          pollingInterval: 8000 // Check every 8 seconds
      });
      await brlt.waitForDeployment();
      const brltAddress = await brlt.getAddress();
      console.log("BRLT (proxy) deployed to:", brltAddress);

      const implementationAddress = await hre.upgrades.erc1967.getImplementationAddress(brltAddress);
      console.log("BRLT implementation deployed to:", implementationAddress);

      if (hre.network.name !== "hardhat" && hre.network.name !== "localhost") {
        console.log("Waiting for block confirmations...");
        // For proxies, the deploymentTransaction might be on the brlt object or need to be fetched differently.
        // Let's assume brlt.deploymentTransaction() is valid for the proxy deployment transaction.
        if (brlt.deploymentTransaction()) {
            await brlt.deploymentTransaction().wait(5);
        } else {
            // If not, a simple delay might be necessary as a fallback before verification
            console.log("Proxy deployment transaction details not directly available, waiting 30s before verification...");
            await new Promise(resolve => setTimeout(resolve, 30000)); 
        }
        
        console.log("Verifying BRLT proxy contract on explorer...");
        // For UUPS proxies, verification is typically done against the proxy address using the implementation's source code.
        // The 'verify:verify' task for proxies might not need constructorArguments if the proxy itself has no constructor args,
        // and the linking to the implementation is handled by the proxy standard.
        // However, Hardhat Upgrades plugin often handles this by verifying the implementation and linking it.
        // Let's try verifying the proxy address directly.
        try {
          await hre.run("verify:verify", {
            address: brltAddress, // Verify the proxy address
            // No constructor arguments needed for the UUPS proxy itself typically
            // If this fails, one might need to verify the implementationAddress separately
            // and Etherscan/Basescan will link them if the proxy points to it.
          });
          console.log("BRLT proxy contract verified successfully (or verification was initiated and linked).");
        } catch (error) {
          if (error.message.toLowerCase().includes("already verified")) {
            console.log(`BRLT proxy contract at ${brltAddress} is already verified/linked.`);
          } else {
            console.error("BRLT proxy verification failed:", error.message);
            console.log("Attempting to verify implementation contract directly as a fallback...");
            try {
              await hre.run("verify:verify", {
                  address: implementationAddress,
                  contract: "solidity/BRLT.sol:BRLT" // Specify contract path for clarity
              });
              console.log(`BRLT implementation contract (${implementationAddress}) verified. Please ensure proxy ${brltAddress} points to it and is manually marked as proxy if needed.`);
            } catch (implError) {
              if (implError.message.toLowerCase().includes("already verified") || implError.message.toLowerCase().includes("has already been verified")) {
                console.log(`BRLT implementation contract at ${implementationAddress} is already verified.`);
              } else {
                console.error("BRLT implementation verification also failed:", implError.message);
              }
            }
          }
        }
      }
      return brltAddress;

    } else {
      console.error(`Unknown or unsupported contract name: ${contractName}. Supported: BRLT.`);
      throw new Error(`Unknown or unsupported contract name: ${contractName}`);
    }
  });

// To make the task available, it needs to be imported in hardhat.config.js
// Example: require("./tasks/deploy-task.js"); 