// SPDX-License-Identifier: MIT
pragma solidity ^0.8.30;

error InsufficientBalance();

contract TestUSDTMock {
    string public constant name = "Test USDT";
    string public constant symbol = "USDT";
    uint8 public constant decimals = 6;

    mapping(address => uint256) public balanceOf;

    event Transfer(address indexed from, address indexed to, uint256 value);

    function mint(address to, uint256 amount) external {
        balanceOf[to] += amount;
        emit Transfer(address(0), to, amount);
    }

    // Intentionally no return value to mimic USDT-style ERC20 compatibility quirks.
    function transfer(address to, uint256 amount) external {
        uint256 senderBalance = balanceOf[msg.sender];
        if (senderBalance < amount) revert InsufficientBalance();

        unchecked {
            balanceOf[msg.sender] = senderBalance - amount;
        }
        balanceOf[to] += amount;

        emit Transfer(msg.sender, to, amount);
    }
}
