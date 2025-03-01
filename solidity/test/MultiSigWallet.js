const { expect } = require("chai");
const { ethers } = require("hardhat");
const { time } = require("@nomicfoundation/hardhat-network-helpers");

describe("MultiSigWallet", function () {
  let wallet;
  let token;
  let manager;
  let client;
  let recovery;
  let other;
  const RECOVERY_DELAY = 72 * 60 * 60; // 72 hours in seconds

  beforeEach(async function () {
    [manager, client, recovery, other] = await ethers.getSigners();

    // Deploy mock token
    const MockToken = await ethers.getContractFactory("MockToken");
    token = await MockToken.deploy();

    // Deploy wallet
    const Wallet = await ethers.getContractFactory("MultiSigWallet");
    wallet = await Wallet.deploy(client.address, recovery.address);

    // Send some ETH to the wallet
    await manager.sendTransaction({
      to: wallet.target,
      value: ethers.parseEther("10.0")
    });

    // Send some tokens to the wallet
    await token.transfer(wallet.target, ethers.parseUnits("100", 18));
  });

  describe("Deployment", function () {
    it("Should set the right manager", async function () {
      expect(await wallet.manager()).to.equal(manager.address);
    });

    it("Should set the right client", async function () {
      expect(await wallet.client()).to.equal(client.address);
    });

    it("Should set the right recovery address", async function () {
      expect(await wallet.recoveryAddress()).to.equal(recovery.address);
    });
  });

  describe("Withdrawals", function () {
    it("Should allow manager to create withdrawal request", async function () {
      const tx = await wallet.requestWithdrawal(
        ethers.ZeroAddress,
        ethers.parseEther("1.0"),
        other.address
      );

      const receipt = await tx.wait();
      const event = receipt.logs.find(
        log => log.fragment && log.fragment.name === 'WithdrawalRequested'
      );
      expect(event).to.not.be.undefined;
    });

    it("Should allow client to create withdrawal request", async function () {
      const tx = await wallet.connect(client).requestWithdrawal(
        ethers.ZeroAddress,
        ethers.parseEther("1.0"),
        other.address
      );

      const receipt = await tx.wait();
      const event = receipt.logs.find(
        log => log.fragment && log.fragment.name === 'WithdrawalRequested'
      );
      expect(event).to.not.be.undefined;
    });

    it("Should not allow non-authorized users to create withdrawal request", async function () {
      await expect(
        wallet.connect(other).requestWithdrawal(
          ethers.ZeroAddress,
          ethers.parseEther("1.0"),
          other.address
        )
      ).to.be.revertedWith("Unauthorized");
    });

    it("Should execute withdrawal after both signatures", async function () {
      const withdrawAmount = ethers.parseEther("1.0");
      const initialBalance = await ethers.provider.getBalance(other.address);

      // Create request
      const tx = await wallet.requestWithdrawal(
        ethers.ZeroAddress,
        withdrawAmount,
        other.address
      );
      const receipt = await tx.wait();
      const requestId = receipt.logs.find(
        log => log.fragment && log.fragment.name === 'WithdrawalRequested'
      ).args[0];

      // Client signs
      await wallet.connect(client).signWithdrawal(requestId);

      // Check balance changed
      const finalBalance = await ethers.provider.getBalance(other.address);
      expect(finalBalance - initialBalance).to.equal(withdrawAmount);
    });
  });

  describe("Token Operations", function () {
    it("Should allow token deposits", async function () {
      const amount = ethers.parseUnits("10", 18);
      await token.approve(wallet.target, amount);
      
      const tx = await wallet.depositToken(token.target, amount);
      const receipt = await tx.wait();
      const event = receipt.logs.find(
        log => log.fragment && log.fragment.name === 'Deposited'
      );
      
      expect(event.args.token).to.equal(token.target);
      expect(event.args.amount).to.equal(amount);
    });

    it("Should not allow token deposits with zero amount", async function () {
      await token.approve(wallet.target, ethers.parseUnits("10", 18));
      await expect(wallet.depositToken(token.target, 0))
        .to.be.revertedWith("Amount must be greater than 0");
    });

    it("Should not allow native coin deposits via depositToken", async function () {
      await expect(wallet.depositToken(ethers.ZeroAddress, ethers.parseEther("1.0")))
        .to.be.revertedWith("Use receive() for native coin");
    });

    it("Should execute token withdrawal after both signatures", async function () {
      const withdrawAmount = ethers.parseUnits("1.0", 18);
      const initialBalance = await token.balanceOf(other.address);

      // Create request
      const tx = await wallet.requestWithdrawal(
        token.target,
        withdrawAmount,
        other.address
      );
      const receipt = await tx.wait();
      const requestId = receipt.logs.find(
        log => log.fragment && log.fragment.name === 'WithdrawalRequested'
      ).args[0];

      // Client signs
      await wallet.connect(client).signWithdrawal(requestId);

      // Check balance changed
      const finalBalance = await token.balanceOf(other.address);
      expect(finalBalance - initialBalance).to.equal(withdrawAmount);
    });
  });

  describe("Withdrawal Edge Cases", function () {
    it("Should not allow withdrawal request after expiration", async function () {
      // Create request
      const tx = await wallet.requestWithdrawal(
        ethers.ZeroAddress,
        ethers.parseEther("1.0"),
        other.address
      );
      const receipt = await tx.wait();
      const requestId = receipt.logs.find(
        log => log.fragment && log.fragment.name === 'WithdrawalRequested'
      ).args[0];

      // Fast forward 25 hours
      await time.increase(25 * 60 * 60);

      // Try to sign expired request
      await expect(wallet.connect(client).signWithdrawal(requestId))
        .to.be.revertedWith("Request expired");
    });

    it("Should not allow withdrawal with zero amount", async function () {
      await expect(wallet.requestWithdrawal(
        ethers.ZeroAddress,
        0,
        other.address
      )).to.be.revertedWith("Invalid amount");
    });

    it("Should not allow withdrawal to zero address", async function () {
      await expect(wallet.requestWithdrawal(
        ethers.ZeroAddress,
        ethers.parseEther("1.0"),
        ethers.ZeroAddress
      )).to.be.revertedWith("Invalid recipient");
    });

    it("Should increment withdrawal nonce correctly", async function () {
      const initialNonce = await wallet.withdrawalNonce();
      
      await wallet.requestWithdrawal(
        ethers.ZeroAddress,
        ethers.parseEther("1.0"),
        other.address
      );

      expect(await wallet.withdrawalNonce()).to.equal(initialNonce + 1n);
    });
  });

  describe("Recovery", function () {
    it("Should allow only manager to request recovery", async function () {
      await expect(wallet.connect(client).requestRecovery())
        .to.be.revertedWith("Only manager can call this function");

      const tx = await wallet.requestRecovery();
      const receipt = await tx.wait();
      const event = receipt.logs.find(
        log => log.fragment && log.fragment.name === 'RecoveryRequested'
      );
      expect(event).to.not.be.undefined;
    });

    it("Should allow only client to cancel recovery within timelock", async function () {
      await wallet.requestRecovery();

      await expect(wallet.cancelRecovery())
        .to.be.revertedWith("Only client can call this function");

      const tx = await wallet.connect(client).cancelRecovery();
      const receipt = await tx.wait();
      const event = receipt.logs.find(
        log => log.fragment && log.fragment.name === 'RecoveryCancelled'
      );
      expect(event).to.not.be.undefined;
    });

    it("Should not allow recovery execution before timelock", async function () {
      await wallet.requestRecovery();

      await expect(wallet.executeRecovery())
        .to.be.revertedWith("Recovery delay not elapsed");
    });

    it("Should allow recovery execution after timelock", async function () {
      const initialEthBalance = await ethers.provider.getBalance(recovery.address);
      const walletEthBalance = await ethers.provider.getBalance(wallet.target);
      const walletTokenBalance = await token.balanceOf(wallet.target);

      // Request recovery
      await wallet.requestRecovery();

      // Fast forward 72 hours
      await time.increase(RECOVERY_DELAY);

      // Execute token recovery first
      await wallet.executeTokenRecovery([token.target]);

      // Check token balances
      expect(await token.balanceOf(recovery.address)).to.equal(walletTokenBalance);
      expect(await token.balanceOf(wallet.target)).to.equal(0);

      // Execute ETH recovery
      await wallet.executeRecovery();

      // Check ETH balances
      const finalEthBalance = await ethers.provider.getBalance(recovery.address);
      expect(finalEthBalance - initialEthBalance).to.equal(walletEthBalance);
      expect(await ethers.provider.getBalance(wallet.target)).to.equal(0);

      // Complete recovery
      await wallet.completeRecovery();

      // Verify recovery is completed
      expect(await wallet.recoveryExecuted()).to.be.true;
      expect(await wallet.recoveryRequestTimestamp()).to.equal(0);
    });

    it("Should not allow recovery execution if cancelled", async function () {
      await wallet.requestRecovery();
      await wallet.connect(client).cancelRecovery();

      // Fast forward 72 hours
      await time.increase(RECOVERY_DELAY);

      await expect(wallet.executeRecovery())
        .to.be.revertedWith("No recovery requested");
    });

    it("Should not allow client to cancel after timelock", async function () {
      await wallet.requestRecovery();

      // Fast forward 72 hours
      await time.increase(RECOVERY_DELAY);

      await expect(wallet.connect(client).cancelRecovery())
        .to.be.revertedWith("Recovery period expired");
    });
  });

  describe("Recovery Edge Cases", function () {
    it("Should not allow requesting recovery when already requested", async function () {
      await wallet.requestRecovery();
      await expect(wallet.requestRecovery())
        .to.be.revertedWith("Recovery already requested");
    });

    it("Should not allow recovery execution when already executed", async function () {
      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);
      await wallet.executeRecovery();
      await wallet.executeTokenRecovery([token.target]);
      await wallet.completeRecovery();

      // Try to request recovery again
      await expect(wallet.requestRecovery())
        .to.be.revertedWith("Recovery already executed");
    });

    it("Should not allow executing recovery without completing previous recovery", async function () {
      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);
      await wallet.executeRecovery();
      await wallet.executeTokenRecovery([token.target]);

      // Try to request recovery again without completing
      await expect(wallet.requestRecovery())
        .to.be.revertedWith("Recovery already requested");
    });

    it("Should not allow executing recovery twice", async function () {
      // Request and execute recovery
      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);
      await wallet.executeRecovery();
      await wallet.completeRecovery();
      
      // Try to execute recovery again without requesting
      await expect(wallet.executeRecovery())
        .to.be.revertedWith("No recovery requested");
    });

    it("Should not allow completing recovery without executing recovery", async function () {
      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);
      
      // Try to complete without executing
      await wallet.executeRecovery();
      await wallet.completeRecovery();

      // Try to complete again
      await expect(wallet.completeRecovery())
        .to.be.revertedWith("No recovery requested");
    });

    it("Should handle empty token recovery array", async function () {
      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);
      
      const tx = await wallet.executeTokenRecovery([]);
      const receipt = await tx.wait();
      
      // Should not emit RecoveryExecuted for empty array
      const events = receipt.logs.filter(
        log => log.fragment && log.fragment.name === 'RecoveryExecuted'
      );
      expect(events.length).to.equal(0);
    });

    it("Should handle empty token recovery", async function () {
      // Transfer all tokens out first
      const balance = await token.balanceOf(wallet.target);
      const network = await ethers.provider.getNetwork();
      await wallet.requestWithdrawal(token.target, balance, other.address);
      await wallet.connect(client).signWithdrawal(
        ethers.keccak256(
          ethers.solidityPacked(
            ["address", "uint256", "address", "uint256", "uint256", "uint256"],
            [token.target, balance, other.address, await time.latest(), 0, network.chainId]
          )
        )
      );

      // Request and execute recovery
      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);
      
      const tx = await wallet.executeTokenRecovery([token.target]);
      const receipt = await tx.wait();
      
      // Should not emit RecoveryExecuted for zero balance
      const events = receipt.logs.filter(
        log => log.fragment && log.fragment.name === 'RecoveryExecuted'
      );
      expect(events.length).to.equal(0);
    });

    it("Should not allow token recovery batch size exceeding limit", async function () {
      const tokens = Array(21).fill(token.target);
      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);

      await expect(wallet.executeTokenRecovery(tokens))
        .to.be.revertedWith("Batch size too large");
    });

    it("Should not allow native coin in token recovery", async function () {
      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);

      await expect(wallet.executeTokenRecovery([ethers.ZeroAddress]))
        .to.be.revertedWith("Use executeRecovery() for native coin");
    });
  });

  describe("Balance Queries", function () {
    it("Should return correct native coin balance", async function () {
      const balance = await wallet.getBalance();
      expect(balance).to.equal(ethers.parseEther("10.0")); // Initial balance from setup
      
      // Send more ETH and check balance updates
      await manager.sendTransaction({
        to: wallet.target,
        value: ethers.parseEther("5.0")
      });
      
      const newBalance = await wallet.getBalance();
      expect(newBalance).to.equal(ethers.parseEther("15.0"));
    });

    it("Should return correct token balance", async function () {
      const balance = await wallet.getTokenBalance(token.target);
      expect(balance).to.equal(ethers.parseUnits("100", 18)); // Initial balance from setup
      
      // Send more tokens and check balance updates
      await token.transfer(wallet.target, ethers.parseUnits("50", 18));
      
      const newBalance = await wallet.getTokenBalance(token.target);
      expect(newBalance).to.equal(ethers.parseUnits("150", 18));
    });

    it("Should not allow getting native coin balance via getTokenBalance", async function () {
      await expect(wallet.getTokenBalance(ethers.ZeroAddress))
        .to.be.revertedWith("Use getBalance() for native coin");
    });

    it("Should handle token balance queries for tokens with no balance", async function () {
      // Deploy a new token that the wallet doesn't have
      const MockToken = await ethers.getContractFactory("MockToken");
      const newToken = await MockToken.deploy();

      const balance = await wallet.getTokenBalance(newToken.target);
      expect(balance).to.equal(0);
    });
  });
});
