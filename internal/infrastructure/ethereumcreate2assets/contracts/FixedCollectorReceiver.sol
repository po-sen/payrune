// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

contract FixedCollectorReceiver {
    error CollectorAddressRequired();
    error SweepFailed();

    address public immutable collector;

    constructor(address collector_) payable {
        if (collector_ == address(0)) {
            revert CollectorAddressRequired();
        }

        collector = collector_;
    }

    receive() external payable {}

    function sweep() external {
        uint256 balance = address(this).balance;
        if (balance == 0) {
            return;
        }

        (bool ok, ) = payable(collector).call{value: balance}("");
        if (!ok) {
            revert SweepFailed();
        }
    }
}
