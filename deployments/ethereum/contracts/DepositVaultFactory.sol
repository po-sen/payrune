// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

interface IERC20Minimal {
    function balanceOf(address account) external view returns (uint256);

    function transfer(address to, uint256 amount) external;
}

error InvalidCollector();
error InvalidToken();
error NativeTransferFailed();
error TokenBalanceQueryFailed();
error TokenTransferFailed();
error Unauthorized();
error VaultDeploymentFailed();

contract DepositVault {
    address public immutable factory;
    address public immutable collector;

    constructor(address collector_) payable {
        if (collector_ == address(0)) revert InvalidCollector();

        factory = msg.sender;
        collector = collector_;
    }

    receive() external payable {}

    function sweepNative() external returns (uint256 amount) {
        if (msg.sender != factory) revert Unauthorized();

        amount = address(this).balance;
        if (amount == 0) {
            return 0;
        }

        (bool ok, ) = payable(collector).call{value: amount}("");
        if (!ok) revert NativeTransferFailed();
    }

    function sweepToken(address token) external returns (uint256 amount) {
        if (msg.sender != factory) revert Unauthorized();
        if (token == address(0)) revert InvalidToken();

        amount = _balanceOf(token);
        if (amount == 0) {
            return 0;
        }

        _safeTransfer(token, collector, amount);
    }

    function _balanceOf(address token) private view returns (uint256 amount) {
        (bool ok, bytes memory data) = token.staticcall(
            abi.encodeWithSelector(IERC20Minimal.balanceOf.selector, address(this))
        );
        if (!ok || data.length < 32) revert TokenBalanceQueryFailed();

        amount = abi.decode(data, (uint256));
    }

    function _safeTransfer(address token, address to, uint256 amount) private {
        (bool ok, bytes memory data) = token.call(
            abi.encodeWithSelector(IERC20Minimal.transfer.selector, to, amount)
        );
        if (!ok) revert TokenTransferFailed();
        if (data.length > 0 && !abi.decode(data, (bool))) revert TokenTransferFailed();
    }
}

contract DepositVaultFactory {
    address public immutable owner;
    address public immutable collector;

    event NativeSwept(bytes32 indexed salt, address indexed vault, uint256 amount);
    event TokenSwept(bytes32 indexed salt, address indexed vault, address indexed token, uint256 amount);
    event VaultDeployed(bytes32 indexed salt, address indexed vault);

    constructor(address collector_) {
        if (collector_ == address(0)) revert InvalidCollector();

        owner = msg.sender;
        collector = collector_;
    }

    modifier onlyOwner() {
        if (msg.sender != owner) revert Unauthorized();
        _;
    }

    function vaultCreationCodeHash() public view returns (bytes32) {
        return keccak256(abi.encodePacked(type(DepositVault).creationCode, abi.encode(collector)));
    }

    function predictVaultAddress(bytes32 salt) public view returns (address predicted) {
        predicted = address(
            uint160(
                uint256(
                    keccak256(
                        abi.encodePacked(
                            bytes1(0xff),
                            address(this),
                            salt,
                            vaultCreationCodeHash()
                        )
                    )
                )
            )
        );
    }

    function deployVault(bytes32 salt) public onlyOwner returns (address vault) {
        vault = predictVaultAddress(salt);
        if (vault.code.length > 0) {
            return vault;
        }

        vault = address(new DepositVault{salt: salt}(collector));
        if (vault != predictVaultAddress(salt)) revert VaultDeploymentFailed();

        emit VaultDeployed(salt, vault);
    }

    function batchDeployAndSweepNative(bytes32[] calldata salts) external onlyOwner {
        for (uint256 i = 0; i < salts.length; i++) {
            address vault = deployVault(salts[i]);
            uint256 amount = DepositVault(payable(vault)).sweepNative();
            emit NativeSwept(salts[i], vault, amount);
        }
    }

    function batchDeployAndSweepToken(bytes32[] calldata salts, address token) external onlyOwner {
        if (token == address(0)) revert InvalidToken();

        for (uint256 i = 0; i < salts.length; i++) {
            address vault = deployVault(salts[i]);
            uint256 amount = DepositVault(payable(vault)).sweepToken(token);
            emit TokenSwept(salts[i], vault, token, amount);
        }
    }
}
