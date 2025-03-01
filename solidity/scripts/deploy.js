const hre = require("hardhat");

async function main() {
  // Get the contract factory
  const MultiSigWallet = await hre.ethers.getContractFactory("MultiSigWallet");

  // Get deployment parameters from environment variables
  const clientAddress = process.env.CLIENT_ADDRESS;
  const recoveryAddress = process.env.RECOVERY_ADDRESS;

  if (!clientAddress || !recoveryAddress) {
    throw new Error("CLIENT_ADDRESS and RECOVERY_ADDRESS must be set in .env file");
  }

  console.log("Deploying MultiSigWallet...");
  console.log("Network:", hre.network.name);
  console.log("Client Address:", clientAddress);
  console.log("Recovery Address:", recoveryAddress);

  // Deploy the contract
  const wallet = await MultiSigWallet.deploy(clientAddress, recoveryAddress);
  await wallet.waitForDeployment();

  const address = await wallet.getAddress();
  console.log("MultiSigWallet deployed to:", address);

  // Wait for a few block confirmations
  console.log("Waiting for block confirmations...");
  await wallet.deploymentTransaction().wait(5);

  // Verify the contract
  if (hre.network.name !== "hardhat" && hre.network.name !== "localhost") {
    console.log("Verifying contract on explorer...");
    try {
      await hre.run("verify:verify", {
        address: address,
        constructorArguments: [clientAddress, recoveryAddress],
      });
      console.log("Contract verified successfully");
    } catch (error) {
      console.log("Verification failed:", error.message);
    }
  }

  return address;
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
