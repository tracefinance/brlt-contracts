// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "@openzeppelin/contracts-upgradeable/token/ERC20/ERC20Upgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/token/ERC20/extensions/ERC20PermitUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/ContextUpgradeable.sol"; // Required for _msgSender()

/**
 * @title BRLT Token (Upgradeable)
 * @author Vault0 Team
 * @notice ERC20 stablecoin pegged to BRL (Brazilian Real). Features include minting, burning,
 * pausing, blacklisting capabilities, and EIP-2612 permit functionality for gasless approvals.
 * @dev Access control is managed via OpenZeppelin's AccessControl. Contract is UUPS upgradeable.
 */
contract BRLT is 
    Initializable, 
    UUPSUpgradeable, 
    ERC20Upgradeable, 
    AccessControlUpgradeable, 
    PausableUpgradeable, 
    ERC20PermitUpgradeable 
{
    /// @dev Role identifier for addresses authorized to mint new tokens.
    bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");
    /// @dev Role identifier for addresses authorized to burn tokens.
    bytes32 public constant BURNER_ROLE = keccak256("BURNER_ROLE");
    /// @dev Role identifier for addresses authorized to pause and unpause the contract.
    bytes32 public constant PAUSER_ROLE = keccak256("PAUSER_ROLE");
    /// @dev Role identifier for addresses authorized to blacklist and unblacklist accounts.
    bytes32 public constant BLACKLISTER_ROLE = keccak256("BLACKLISTER_ROLE");
    /// @dev Role identifier for addresses authorized to upgrade the contract.
    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");

    // Mapping to store blacklisted addresses
    mapping(address => bool) private _isBlacklisted;

    /**
     * @notice Emitted when an `account`'s blacklist status is changed.
     * @param account The address whose blacklist status was changed.
     * @param isBlacklisted The new blacklist status of the account (true if blacklisted, false otherwise).
     */
    event AddressBlacklistedStatusChanged(address indexed account, bool isBlacklisted);

    // Custom error for blacklisted accounts, to be used in _update
    error AccountBlacklisted(address account);

    /**
     * @dev Initializer for the BRLT token contract.
     * Sets the token name to "BRLT" and symbol to "BRLT".
     * Initializes ERC20Permit with the token name "BRLT".
     * Grants the `initialAdmin` address the `DEFAULT_ADMIN_ROLE` and all operational roles
     * (MINTER_ROLE, BURNER_ROLE, PAUSER_ROLE, BLACKLISTER_ROLE, UPGRADER_ROLE).
     * @param initialAdmin The address to receive all initial roles.
     */
    function initialize(address initialAdmin) public virtual initializer {
        require(initialAdmin != address(0), "BRLT: initial admin cannot be zero address");

        __ERC20_init("BRLT", "BRLT");
        __ERC20Permit_init("BRLT");
        __AccessControl_init();
        __Pausable_init();
        __UUPSUpgradeable_init();

        _grantRole(DEFAULT_ADMIN_ROLE, initialAdmin);
        _grantRole(MINTER_ROLE, initialAdmin);
        _grantRole(BURNER_ROLE, initialAdmin);
        _grantRole(PAUSER_ROLE, initialAdmin);
        _grantRole(BLACKLISTER_ROLE, initialAdmin);
        _grantRole(UPGRADER_ROLE, initialAdmin);
    }

    /**
     * @notice Returns the number of decimals used to get its user representation (18).
     * @return The number of decimals (18).
     */
    function decimals() public pure override(ERC20Upgradeable) returns (uint8) {
        return 6;
    }

    /**
     * @dev See {IERC165-supportsInterface}.
     */
    function supportsInterface(bytes4 interfaceId) public view virtual override(AccessControlUpgradeable) returns (bool) {
        return interfaceId == 0xd505accf || super.supportsInterface(interfaceId); // 0xd505accf is IERC20Permit's ID
    }
    
    /**
     * @dev Hook that is called before any transfer of tokens. This includes minting and burning.
     * Ensures that the contract is not paused and that neither the sender (`from`) nor the
     * receiver (`to`) are blacklisted.
     */
    function _update(address from, address to, uint256 value) internal virtual override(ERC20Upgradeable) {
        _requireNotPaused(); // Explicit pause check

        if (from != address(0) && _isBlacklisted[from]) {
            revert AccountBlacklisted(from);
        }
        if (to != address(0) && _isBlacklisted[to]) {
            revert AccountBlacklisted(to);
        }
        super._update(from, to, value); // Calls ERC20Upgradeable._update
    }

    /**
     * @notice Mints `amount` new tokens and assigns them to `account`.
     */
    function mint(address account, uint256 amount) external virtual onlyRole(MINTER_ROLE) {
        _requireNotPaused(); // Explicit pause check
        _mint(account, amount);
    }

    /**
     * @notice Burns `amount` tokens from `account`.
     */
    function burnFrom(address account, uint256 amount) external virtual onlyRole(BURNER_ROLE) {
        _requireNotPaused(); // Explicit pause check
        _burn(account, amount);
    }

    /**
     * @notice Pauses all token transfers, minting, and burning operations.
     */
    function pause() external virtual onlyRole(PAUSER_ROLE) {
        _pause();
    }

    /**
     * @notice Unpauses all token transfers, minting, and burning operations.
     */
    function unpause() external virtual onlyRole(PAUSER_ROLE) {
        _unpause();
    }

    /**
     * @notice Adds `account` to the blacklist.
     */
    function blacklistAddress(address account) external virtual onlyRole(BLACKLISTER_ROLE) {
        require(account != address(0), "BRLT: cannot blacklist zero address");
        _isBlacklisted[account] = true;
        emit AddressBlacklistedStatusChanged(account, true);
    }

    /**
     * @notice Removes `account` from the blacklist.
     */
    function unblacklistAddress(address account) external virtual onlyRole(BLACKLISTER_ROLE) {
        _isBlacklisted[account] = false;
        emit AddressBlacklistedStatusChanged(account, false);
    }

    /**
     * @notice Checks if an `account` is currently blacklisted.
     */
    function isBlacklisted(address account) public view returns (bool) {
        return _isBlacklisted[account];
    }

    /**
     * @dev Overrides ERC20PermitUpgradeable's permit function to include Pausable check.
     * @inheritdoc ERC20PermitUpgradeable
     */
    function permit(address owner, address spender, uint256 value, uint256 deadline, uint8 v, bytes32 r, bytes32 s) public virtual override(ERC20PermitUpgradeable) {
        _requireNotPaused(); // Explicit pause check (already correctly here)
        super.permit(owner, spender, value, deadline, v, r, s);
    }

    /**
     * @dev Hook for UUPS upgrades. Authorizes an upgrade if the caller has the `UPGRADER_ROLE`.
     * @param newImplementation The address of the new implementation contract.
     */
    function _authorizeUpgrade(address newImplementation) internal virtual override onlyRole(UPGRADER_ROLE) {
        // This function body being empty means onlyRole(UPGRADER_ROLE) check is sufficient.
        // Custom logic can be added here if needed beyond role check.
    }
} 