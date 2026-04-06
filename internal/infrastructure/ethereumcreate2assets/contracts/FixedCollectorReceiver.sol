// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import "./Create2RecoveryInterfaces.sol";

contract FixedCollectorReceiver is IFixedCollectorERC20Receiver {
    error CollectorAddressRequired();
    error TokenAddressRequired();
    error SweepFailed();
    error TokenSweepFailed();

    address public immutable override collector;

    constructor(address collector_) payable {
        if (collector_ == address(0)) {
            revert CollectorAddressRequired();
        }

        collector = collector_;
    }

    receive() external payable {}

    function sweep() external override {
        uint256 balance = address(this).balance;
        if (balance == 0) {
            return;
        }

        (bool ok, ) = payable(collector).call{value: balance}("");
        if (!ok) {
            revert SweepFailed();
        }
    }

    function sweepERC20(address token) external override {
        if (token == address(0)) {
            revert TokenAddressRequired();
        }

        uint256 balance = IERC20BalanceReader(token).balanceOf(address(this));
        if (balance == 0) {
            return;
        }

        (bool ok, bytes memory result) = token.call(
            abi.encodeWithSignature("transfer(address,uint256)", collector, balance)
        );
        if (!ok) {
            revert TokenSweepFailed();
        }
        if (result.length > 0 && !abi.decode(result, (bool))) {
            revert TokenSweepFailed();
        }
    }
}
