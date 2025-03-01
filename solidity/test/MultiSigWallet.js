const { expect } = require("chai");
const { ethers } = require("hardhat");
const { time } = require("@nomicfoundation/hardhat-network-helpers");

describe("MultiSigWallet", function () {
  let wallet;
  let token;
  let token2; 
  let manager;
  let client;
  let recovery;
  let other;
  const RECOVERY_DELAY = 72 * 60 * 60; 

  beforeEach(async function () {
    [manager, client, recovery, other] = await ethers.getSigners();

    const MockToken = await ethers.getContractFactory("contracts/mocks/MockToken.sol:MockToken");
    token = await MockToken.deploy();
    token2 = await MockToken.deploy(); 

    const Wallet = await ethers.getContractFactory("MultiSigWallet");
    wallet = await Wallet.deploy(client.address, recovery.address);

    await manager.sendTransaction({
      to: wallet.target,
      value: ethers.parseEther("10.0")
    });

    await token.transfer(wallet.target, ethers.parseUnits("100", 18));
    await token2.transfer(wallet.target, ethers.parseUnits("50", 18));
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

      const tx = await wallet.requestWithdrawal(
        ethers.ZeroAddress,
        withdrawAmount,
        other.address
      );
      const receipt = await tx.wait();
      const requestId = receipt.logs.find(
        log => log.fragment && log.fragment.name === 'WithdrawalRequested'
      ).args[0];

      await wallet.connect(client).signWithdrawal(requestId);

      const finalBalance = await ethers.provider.getBalance(other.address);
      expect(finalBalance - initialBalance).to.equal(withdrawAmount);
    });
    
    it("Should not allow withdrawal request during recovery mode", async function () {
      await wallet.requestRecovery();
      
      await expect(
        wallet.requestWithdrawal(
          ethers.ZeroAddress,
          ethers.parseEther("1.0"),
          other.address
        )
      ).to.be.revertedWith("Wallet in recovery mode");
      
      await expect(
        wallet.connect(manager).requestWithdrawal(
          token.target,
          ethers.parseUnits("1.0", 18),
          other.address
        )
      ).to.be.revertedWith("Wallet in recovery mode");
    });
    
    it("Should allow withdrawal request after recovery is cancelled", async function () {
      await wallet.requestRecovery();
      
      await wallet.connect(client).cancelRecovery();
      
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

      const tx = await wallet.requestWithdrawal(
        token.target,
        withdrawAmount,
        other.address
      );
      const receipt = await tx.wait();
      const requestId = receipt.logs.find(
        log => log.fragment && log.fragment.name === 'WithdrawalRequested'
      ).args[0];

      await wallet.connect(client).signWithdrawal(requestId);

      const finalBalance = await token.balanceOf(other.address);
      expect(finalBalance - initialBalance).to.equal(withdrawAmount);
    });
    
    it("Should not allow token deposits during recovery mode", async function () {
      const amount = ethers.parseUnits("10", 18);
      await token.approve(wallet.target, amount);
      
      await wallet.requestRecovery();
      
      await expect(
        wallet.depositToken(token.target, amount)
      ).to.be.revertedWith("Wallet in recovery mode");
    });
    
    it("Should not allow ETH deposits during recovery mode", async function () {
      await wallet.requestRecovery();
      
      await expect(
        manager.sendTransaction({
          to: wallet.target,
          value: ethers.parseEther("1.0")
        })
      ).to.be.revertedWith("Wallet in recovery mode");
    });
  });

  describe("Withdrawal Edge Cases", function () {
    it("Should not allow withdrawal request after expiration", async function () {
      const tx = await wallet.requestWithdrawal(
        ethers.ZeroAddress,
        ethers.parseEther("1.0"),
        other.address
      );
      const receipt = await tx.wait();
      const requestId = receipt.logs.find(
        log => log.fragment && log.fragment.name === 'WithdrawalRequested'
      ).args[0];

      await time.increase(25 * 60 * 60);

      await expect(wallet.connect(client).signWithdrawal(requestId))
        .to.be.revertedWith("Request expired");
    });

    it("Should not execute withdrawal if request expired after both signatures", async function () {
      const MockWallet = await ethers.getContractFactory("MockMultiSigWalletTest");
      const mockWallet = await MockWallet.deploy(client.address, recovery.address);
      
      await manager.sendTransaction({
        to: mockWallet.target,
        value: ethers.parseEther("5.0")
      });
      
      const tx = await mockWallet.requestWithdrawal(
        ethers.ZeroAddress,
        ethers.parseEther("1.0"),
        other.address
      );
      const receipt = await tx.wait();
      const requestId = receipt.logs.find(
        log => log.fragment && log.fragment.name === 'WithdrawalRequested'
      ).args[0];
      
      await mockWallet.connect(client).signWithdrawalWithoutExecution(requestId);
      
      await time.increase(25 * 60 * 60);
      
      await expect(mockWallet.executeWithdrawalDirect(requestId))
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

    it("Should not allow withdrawal with insufficient token balance", async function () {
      // Get current token balance
      const currentBalance = await token.balanceOf(wallet.target);
      
      // Create a withdrawal request for more tokens than available
      const excessAmount = currentBalance + ethers.parseUnits("10", 18);
      
      // Request withdrawal for an excessive amount
      const tx = await wallet.requestWithdrawal(
        token.target,
        excessAmount,
        other.address
      );
      
      const receipt = await tx.wait();
      const requestId = receipt.logs.find(
        log => log.fragment && log.fragment.name === 'WithdrawalRequested'
      ).args[0];
      
      // Try to complete the withdrawal with client signature
      // This should fail with "Insufficient token balance"
      await expect(wallet.connect(client).signWithdrawal(requestId))
        .to.be.revertedWith("Insufficient token balance");
    });

    it("Should not allow direct withdrawal execution with insufficient token balance", async function () {
      const MockWallet = await ethers.getContractFactory("MockMultiSigWalletTest");
      const mockWallet = await MockWallet.deploy(client.address, recovery.address);
      
      // Transfer some tokens to the mock wallet
      await token.transfer(mockWallet.target, ethers.parseUnits("5", 18));
      
      // Create a withdrawal request for more tokens than available
      const tx = await mockWallet.requestWithdrawal(
        token.target,
        ethers.parseUnits("10", 18),  // More than the 5 available
        other.address
      );
      
      const receipt = await tx.wait();
      const requestId = receipt.logs.find(
        log => log.fragment && log.fragment.name === 'WithdrawalRequested'
      ).args[0];
      
      // Sign without executing
      await mockWallet.connect(client).signWithdrawalWithoutExecution(requestId);
      
      // Try to execute directly - should fail with insufficient balance
      await expect(mockWallet.executeWithdrawalDirect(requestId))
        .to.be.revertedWith("Insufficient token balance");
    });

    it("Should generate different requestIds for identical parameters from different signers", async function () {
      const amount = ethers.parseEther("1.0");
      const recipient = other.address;
      
      // Create first withdrawal request from manager
      const managerTx = await wallet.requestWithdrawal(
        ethers.ZeroAddress,
        amount,
        recipient
      );
      const managerReceipt = await managerTx.wait();
      const managerRequestId = managerReceipt.logs.find(
        log => log.fragment && log.fragment.name === 'WithdrawalRequested'
      ).args[0];
      
      // Create second withdrawal request with identical parameters from client
      const clientTx = await wallet.connect(client).requestWithdrawal(
        ethers.ZeroAddress,
        amount,
        recipient
      );
      const clientReceipt = await clientTx.wait();
      const clientRequestId = clientReceipt.logs.find(
        log => log.fragment && log.fragment.name === 'WithdrawalRequested'
      ).args[0];
      
      // Verify requestIds are different despite identical parameters
      expect(managerRequestId).to.not.equal(clientRequestId);
      
      // Verify both requests can be processed independently
      await wallet.connect(client).signWithdrawal(managerRequestId);
      
      // Create a new withdrawal for the client to sign
      const newTx = await wallet.requestWithdrawal(
        ethers.ZeroAddress,
        amount,
        recipient
      );
      const newReceipt = await newTx.wait();
      const newRequestId = newReceipt.logs.find(
        log => log.fragment && log.fragment.name === 'WithdrawalRequested'
      ).args[0];
      
      // Verify the new requestId is also different
      expect(newRequestId).to.not.equal(managerRequestId);
      expect(newRequestId).to.not.equal(clientRequestId);
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

      await wallet.addSupportedToken(token.target);

      await wallet.requestRecovery();

      await time.increase(RECOVERY_DELAY);

      await wallet.executeRecovery();

      expect(await token.balanceOf(recovery.address)).to.equal(walletTokenBalance);
      expect(await token.balanceOf(wallet.target)).to.equal(0);

      const finalEthBalance = await ethers.provider.getBalance(recovery.address);
      expect(finalEthBalance - initialEthBalance).to.equal(walletEthBalance);
      expect(await ethers.provider.getBalance(wallet.target)).to.equal(0);

      expect(await wallet.recoveryExecuted()).to.be.true;
      expect(await wallet.recoveryRequestTimestamp()).to.equal(0);
    });

    it("Should not allow recovery execution if cancelled", async function () {
      await wallet.requestRecovery();
      await wallet.connect(client).cancelRecovery();

      await time.increase(RECOVERY_DELAY);

      await expect(wallet.executeRecovery())
        .to.be.revertedWith("No recovery requested");
    });

    it("Should not allow client to cancel after timelock", async function () {
      await wallet.requestRecovery();

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

      await expect(wallet.requestRecovery())
        .to.be.revertedWith("Recovery already executed");
    });

    it("Should not allow executing recovery without completing previous recovery", async function () {
      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);
      await wallet.executeRecovery();

      await expect(wallet.requestRecovery())
        .to.be.revertedWith("Recovery already executed");
    });

    it("Should not allow executing recovery twice", async function () {
      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);
      await wallet.executeRecovery();

      await expect(wallet.executeRecovery())
        .to.be.revertedWith("No recovery requested");
    });

    it("Should handle empty token recovery", async function () {
      // Set up contract with no tokens to recover
      const tx = await wallet.requestWithdrawal(token.target, await token.balanceOf(wallet.target), other.address);
      const receipt = await tx.wait();
      const requestId = receipt.logs.find(
        log => log.fragment && log.fragment.name === 'WithdrawalRequested'
      ).args[0];
      
      await wallet.connect(client).signWithdrawal(requestId);

      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);
      
      // First add the token as supported
      await wallet.addSupportedToken(token.target);
      
      // Then execute recovery
      const recoveryTx = await wallet.executeRecovery();
      const recoveryReceipt = await recoveryTx.wait();
      
      // Should not emit RecoveryExecuted for token with zero balance
      const tokenEvents = recoveryReceipt.logs.filter(
        log => log.fragment && log.fragment.name === 'RecoveryExecuted' && 
        log.args && log.args[0] === token.target
      );
      expect(tokenEvents.length).to.equal(0);
    });
    
    it("Should not allow non-supported tokens to be recovered in executeRecovery", async function () {
      // Add token1 but not token2
      await wallet.addSupportedToken(token.target);
      
      const initialToken2Balance = await token2.balanceOf(wallet.target);
      
      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);
      
      await wallet.executeRecovery();
      
      // Verify token2 was not recovered
      expect(await token2.balanceOf(recovery.address)).to.equal(0);
      expect(await token2.balanceOf(wallet.target)).to.equal(initialToken2Balance);
    });
  });

  describe("Supported Tokens Management", function () {
    it("Should set native coin as supported by default", async function () {
      expect(await wallet.supportedTokens(ethers.ZeroAddress)).to.be.true;
    });

    it("Should allow manager to add supported tokens", async function () {
      expect(await wallet.supportedTokens(token.target)).to.be.false;
      
      const tx = await wallet.addSupportedToken(token.target);
      const receipt = await tx.wait();
      
      const event = receipt.logs.find(log => log.fragment && log.fragment.name === 'TokenSupported');
      expect(event).to.not.be.undefined;
      expect(event.args.token).to.equal(token.target);
      
      expect(await wallet.supportedTokens(token.target)).to.be.true;
      
      // Verify it was added to the list
      const supportedTokens = await wallet.getSupportedTokens();
      expect(supportedTokens).to.include(token.target);
    });

    it("Should not allow adding a token that is already supported", async function () {
      await wallet.addSupportedToken(token.target);
      await expect(wallet.addSupportedToken(token.target))
        .to.be.revertedWith("Token already supported");
    });

    it("Should allow manager to remove supported tokens", async function () {
      await wallet.addSupportedToken(token.target);
      expect(await wallet.supportedTokens(token.target)).to.be.true;
      
      const tx = await wallet.removeSupportedToken(token.target);
      const receipt = await tx.wait();
      
      const event = receipt.logs.find(log => log.fragment && log.fragment.name === 'TokenRemoved');
      expect(event).to.not.be.undefined;
      expect(event.args.token).to.equal(token.target);
      
      expect(await wallet.supportedTokens(token.target)).to.be.false;
      
      // Verify it was removed from the list
      const supportedTokens = await wallet.getSupportedTokens();
      expect(supportedTokens).to.not.include(token.target);
    });

    it("Should not allow removing a token that is not supported", async function () {
      await expect(wallet.removeSupportedToken(token.target))
        .to.be.revertedWith("Token not in supported list");
    });

    it("Should not allow non-manager to add supported tokens", async function () {
      await expect(wallet.connect(client).addSupportedToken(token.target))
        .to.be.revertedWith("Only manager can call this function");
    });

    it("Should not allow non-manager to remove supported tokens", async function () {
      await wallet.addSupportedToken(token.target);
      await expect(wallet.connect(client).removeSupportedToken(token.target))
        .to.be.revertedWith("Only manager can call this function");
    });
    
    it("Should correctly maintain the supported tokens list", async function () {
      // Add several tokens
      await wallet.addSupportedToken(token.target);
      await wallet.addSupportedToken(token2.target);
      
      // Get the list
      let supportedTokens = await wallet.getSupportedTokens();
      expect(supportedTokens.length).to.equal(3); // ETH + 2 tokens
      expect(supportedTokens).to.include(ethers.ZeroAddress);
      expect(supportedTokens).to.include(token.target);
      expect(supportedTokens).to.include(token2.target);
      
      // Remove a token
      await wallet.removeSupportedToken(token.target);
      
      // Check list again
      supportedTokens = await wallet.getSupportedTokens();
      expect(supportedTokens.length).to.equal(2); // ETH + token2
      expect(supportedTokens).to.include(ethers.ZeroAddress);
      expect(supportedTokens).to.not.include(token.target);
      expect(supportedTokens).to.include(token2.target);
    });
  });

  describe("Modified Recovery Process", function () {
    it("Should only recover supported tokens during recovery", async function () {
      const walletToken1Balance = await token.balanceOf(wallet.target);
      const walletToken2Balance = await token2.balanceOf(wallet.target);
      
      await wallet.addSupportedToken(token.target);
      
      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);
      
      await wallet.executeRecovery();
      
      expect(await token.balanceOf(recovery.address)).to.equal(walletToken1Balance);
      expect(await token.balanceOf(wallet.target)).to.equal(0);
      
      expect(await token2.balanceOf(recovery.address)).to.equal(0);
      expect(await token2.balanceOf(wallet.target)).to.equal(walletToken2Balance);
    });

    it("Should not recover ETH if it's removed from supported tokens", async function () {
      const initialEthBalance = await ethers.provider.getBalance(recovery.address);
      const walletEthBalance = await ethers.provider.getBalance(wallet.target);
      
      await wallet.removeSupportedToken(ethers.ZeroAddress);
      
      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);
      
      await wallet.executeRecovery();
      
      const finalEthBalance = await ethers.provider.getBalance(recovery.address);
      expect(finalEthBalance).to.equal(initialEthBalance);
      expect(await ethers.provider.getBalance(wallet.target)).to.equal(walletEthBalance);
    });

    it("Should allow manager to recover non-supported tokens after recovery is complete", async function () {
      const walletToken2Balance = await token2.balanceOf(wallet.target);
      
      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);
      
      await wallet.executeRecovery();
      
      const tx = await wallet.recoverNonSupportedToken(token2.target, other.address);
      const receipt = await tx.wait();
      
      const event = receipt.logs.find(log => log.fragment && log.fragment.name === 'NonSupportedTokenRecovered');
      expect(event).to.not.be.undefined;
      expect(event.args.token).to.equal(token2.target);
      expect(event.args.amount).to.equal(walletToken2Balance);
      expect(event.args.to).to.equal(other.address);
      
      expect(await token2.balanceOf(other.address)).to.equal(walletToken2Balance);
      expect(await token2.balanceOf(wallet.target)).to.equal(0);
    });

    it("Should not allow recovering non-supported tokens before recovery is complete", async function () {
      await expect(wallet.recoverNonSupportedToken(token2.target, other.address))
        .to.be.revertedWith("Recovery not completed");
        
      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);
      
      await expect(wallet.recoverNonSupportedToken(token2.target, other.address))
        .to.be.revertedWith("Recovery not completed");
    });

    it("Should not allow recovering supported tokens with recoverNonSupportedToken", async function () {
      await wallet.addSupportedToken(token.target);
      
      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);
      
      await wallet.executeRecovery();
      
      await expect(wallet.recoverNonSupportedToken(token.target, other.address))
        .to.be.revertedWith("Use regular recovery for supported tokens");
    });

    it("Should not allow recovering native coin with recoverNonSupportedToken", async function () {
      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);
      
      await wallet.executeRecovery();
      
      await expect(wallet.recoverNonSupportedToken(ethers.ZeroAddress, other.address))
        .to.be.revertedWith("Cannot recover native coin");
    });

    it("Should require a valid recipient address for recoverNonSupportedToken", async function () {
      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);
      
      await wallet.executeRecovery();
      
      await expect(wallet.recoverNonSupportedToken(token2.target, ethers.ZeroAddress))
        .to.be.revertedWith("Invalid recipient address");
    });

    it("Should not allow non-manager to recover non-supported tokens", async function () {
      await wallet.requestRecovery();
      await time.increase(RECOVERY_DELAY);
      
      await wallet.executeRecovery();
      
      await expect(wallet.connect(client).recoverNonSupportedToken(token2.target, other.address))
        .to.be.revertedWith("Only manager can call this function");
    });
  });

  describe("Balance Queries", function () {
    it("Should return correct native coin balance", async function () {
      const balance = await wallet.getBalance();
      expect(balance).to.equal(ethers.parseEther("10.0")); 
      
      await manager.sendTransaction({
        to: wallet.target,
        value: ethers.parseEther("5.0")
      });
      
      const newBalance = await wallet.getBalance();
      expect(newBalance).to.equal(ethers.parseEther("15.0"));
    });

    it("Should return correct token balance", async function () {
      const balance = await wallet.getTokenBalance(token.target);
      expect(balance).to.equal(ethers.parseUnits("100", 18)); 
      
      await token.transfer(wallet.target, ethers.parseUnits("50", 18));
      
      const newBalance = await wallet.getTokenBalance(token.target);
      expect(newBalance).to.equal(ethers.parseUnits("150", 18));
    });

    it("Should not allow getting native coin balance via getTokenBalance", async function () {
      await expect(wallet.getTokenBalance(ethers.ZeroAddress))
        .to.be.revertedWith("Use getBalance() for native coin");
    });

    it("Should handle token balance queries for tokens with no balance", async function () {
      const MockToken = await ethers.getContractFactory("contracts/mocks/MockToken.sol:MockToken");
      const newToken = await MockToken.deploy();

      const balance = await wallet.getTokenBalance(newToken.target);
      expect(balance).to.equal(0);
    });
  });
});
