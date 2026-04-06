// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

interface IFixedCollectorReceiver {
    function collector() external view returns (address);

    function sweep() external;
}

interface IFixedCollectorERC20Receiver is IFixedCollectorReceiver {
    function sweepERC20(address token) external;
}

interface ICreate2SweepFactory {
    function sweep(bytes32[] calldata salts, bytes[] calldata initCodes) external;

    function sweepERC20(
        bytes32[] calldata salts,
        bytes[] calldata initCodes,
        address token
    ) external;
}

interface IERC20BalanceReader {
    function balanceOf(address account) external view returns (uint256);
}
