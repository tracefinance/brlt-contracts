// This setup uses Hardhat Ignition to manage smart contract deployments.
// Learn more about it at https://hardhat.org/ignition

const { buildModule } = require("@nomicfoundation/hardhat-ignition/modules");

module.exports = buildModule("MultiSigWalletModule", (m) => {
  // Get deployment parameters with defaults
  const client = m.getParameter("client", "0x0000000000000000000000000000000000000000");
  const recoveryAddress = m.getParameter("recoveryAddress", "0x0000000000000000000000000000000000000000");

  // Deploy the wallet contract
  const wallet = m.contract("MultiSigWallet", [client, recoveryAddress]);

  // For testing purposes, also deploy the mock token
  const mockToken = m.contract("MockToken", []);

  return { wallet, mockToken };
});
