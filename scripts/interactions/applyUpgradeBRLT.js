const { task } = require("hardhat/config");

task("apply-upgrade-brlt", "Applies an upgrade to the BRLT contract by pointing the proxy to a new implementation")
  .addParam("proxy", "Address of the BRLT proxy contract")
  .addParam("newImplAddress", "Address of the new BRLT implementation contract already deployed")
  .setAction(async (taskArgs, hre) => {
    const { ethers, network, upgrades } = hre;
    const [upgraderSigner] = await ethers.getSigners(); // Signer must have UPGRADER_ROLE

    const { proxy: proxyAddress, newImplAddress } = taskArgs;

    if (!ethers.isAddress(proxyAddress)) {
      console.error(`Error: Invalid proxy address provided: ${proxyAddress}`);
      throw new Error("Invalid proxy address");
    }
    if (!ethers.isAddress(newImplAddress)) {
      console.error(`Error: Invalid new implementation address provided: ${newImplAddress}`);
      throw new Error("Invalid new implementation address");
    }

    console.log(`Executing apply-upgrade-brlt task on network ${network.name}`);
    console.log(`BRLT Proxy Address: ${proxyAddress}`);
    console.log(`New Implementation Address: ${newImplAddress}`);
    console.log(`Signer (upgrader): ${upgraderSigner.address}\n`);

    // Minimal ABI for IERC1967's upgradeTo function
    // const IERC1967_ABI = [
    //     "function upgradeTo(address newImplementation) external",
    //     // Also include an event to potentially listen to, though not strictly needed for the call
    //     "event Upgraded(address indexed implementation)"
    // ];

    try {
        const BRLTFactory = await ethers.getContractFactory("BRLT", upgraderSigner);
        const brltProxy = BRLTFactory.attach(proxyAddress);

        // Log all function fragments from the BRLT contract's ABI
        // console.log("\nBRLT Contract ABI Fragments (Functions):");
        // BRLTFactory.interface.fragments
        //     .filter(fragment => fragment.type === "function")
        //     .forEach(func => console.log(`  - ${func.name}(${func.inputs.map(i => i.type).join(',')})`));
        // console.log("\n");

        // Check UPGRADER_ROLE (optional here, as the upgradeTo call itself will fail if unauthorized)
        // but good for a preemptive check.
        const upgraderRole = await brltProxy.UPGRADER_ROLE();
        const hasUpgraderRole = await brltProxy.hasRole(upgraderRole, upgraderSigner.address);
        if (!hasUpgraderRole) {
            console.error(`Error: Signer ${upgraderSigner.address} does not have the UPGRADER_ROLE on proxy ${proxyAddress}.`);
            throw new Error(`Signer ${upgraderSigner.address} does not have UPGRADER_ROLE`);
        }
        console.log(`Signer ${upgraderSigner.address} has UPGRADER_ROLE. Proceeding with upgrade call.`);

        // Check current implementation to see if upgrade is even needed
        const currentImpl = await upgrades.erc1967.getImplementationAddress(proxyAddress);
        console.log(`Current implementation address: ${currentImpl}`);
        if (currentImpl.toLowerCase() === newImplAddress.toLowerCase()) {
            console.log(`Proxy ${proxyAddress} is already pointing to implementation ${newImplAddress}. No upgrade needed.`);
            return;
        }

        console.log(`Attempting to call upgradeTo(${newImplAddress}) on proxy ${proxyAddress}...`);
        // The BRLT contract (as a UUPS proxy) should have an upgradeTo function.
        // We'll use the IERC1967 ABI to make sure the function is callable.
        // const proxyForUpgrade = new ethers.Contract(proxyAddress, IERC1967_ABI, upgraderSigner);
        
        // Call upgradeToAndCall as it's present in the ABI
        const emptyData = ethers.getBytes("0x");
        const tx = await brltProxy.connect(upgraderSigner).upgradeToAndCall(newImplAddress, emptyData);
        console.log(`Upgrade transaction sent: ${tx.hash}`);
        await tx.wait();
        console.log("Upgrade transaction confirmed.");

        const newCurrentImplementation = await upgrades.erc1967.getImplementationAddress(proxyAddress);
        console.log(`Proxy ${proxyAddress} now points to implementation: ${newCurrentImplementation}`);

        if (newCurrentImplementation.toLowerCase() !== newImplAddress.toLowerCase()) {
            throw new Error(`Upgrade failed: Proxy implementation address ${newCurrentImplementation} does not match expected ${newImplAddress}.`);
        } else {
            console.log(`Successfully upgraded proxy ${proxyAddress} to implementation ${newImplAddress}.`);
        }
        
        // If the new implementation has an initializer (e.g. initializeV2, initializeV3)
        // it should be called here. The `brltProxy` instance might need to be re-attached 
        // with the ABI of the new implementation if the function signature is not in BRLT.sol's ABI.
        // For example, if BRLTv2 has initializeV2():
        // const NewImplFactory = await ethers.getContractFactory("BRLTv2"); // Or whatever newImplName was
        // const newImplInstance = NewImplFactory.attach(proxyAddress).connect(upgraderSigner);
        // if (typeof newImplInstance.initializeV2 === "function") { 
        //    console.log("Calling initializeV2()...");
        //    await newImplInstance.initializeV2(); 
        //    console.log("initializeV2 called.");
        // }

    } catch (error) {
        console.error(`Error during apply upgrade to ${newImplAddress}:`, error);
        throw error;
    }
  }); 