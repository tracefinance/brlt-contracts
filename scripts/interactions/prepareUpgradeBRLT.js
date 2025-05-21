const { task } = require("hardhat/config");

task("prepare-upgrade-brlt", "Prepares an upgrade for the BRLT contract by deploying the new implementation")
  .addParam("proxy", "Address of the BRLT proxy contract")
  .addParam("newImplName", "The name of the new implementation contract (e.g., BRLTv2, BRLTv3)")
  .setAction(async (taskArgs, hre) => {
    const { ethers, network, upgrades } = hre;
    const { proxy: proxyAddress, newImplName } = taskArgs;

    if (!ethers.isAddress(proxyAddress)) {
      console.error(`Error: Invalid proxy address provided: ${proxyAddress}`);
      throw new Error("Invalid proxy address");
    }
    if (!newImplName || typeof newImplName !== 'string') {
        console.error(`Error: New implementation name must be a non-empty string.`);
        throw new Error("Invalid new implementation name");
    }

    console.log(`Executing prepare-upgrade-brlt task on network ${network.name}`);
    console.log(`BRLT Proxy Address: ${proxyAddress}`);
    console.log(`New Implementation Contract Name: ${newImplName}\n`);

    try {
        console.log(`Fetching new implementation contract factory for ${newImplName}...`);
        const NewImplementationFactory = await ethers.getContractFactory(newImplName);
        
        console.log(`Preparing upgrade for proxy ${proxyAddress} to ${newImplName}...`);
        const newImplementationAddress = await upgrades.prepareUpgrade(proxyAddress, NewImplementationFactory, {
            kind: 'uups' // Ensure UUPS is specified if not default or obvious from proxy
        });

        console.log(`New implementation for ${newImplName} deployed at: ${newImplementationAddress}`);
        console.log(`Proxy ${proxyAddress} is now ready to be upgraded to use this implementation.`);
        console.log(`To complete the upgrade, run the 'apply-upgrade-brlt' task with the new implementation address or have an UPGRADER_ROLE holder call upgradeTo().`);

    } catch (error) {
        console.error(`Error during prepare upgrade for ${newImplName}:`, error);
        throw error;
    }
  }); 