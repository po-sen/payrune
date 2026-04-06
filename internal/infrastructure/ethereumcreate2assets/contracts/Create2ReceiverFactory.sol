// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import "./Create2RecoveryInterfaces.sol";

contract Create2ReceiverFactory is ICreate2SweepFactory {
    error Create2DeployFailed();
    error RecoveryInputsRequired();
    error RecoveryInputsLengthMismatch();
    error InitCodeRequired(uint256 index);
    error SweepCallFailed(uint256 index, address receiver);
    error TokenAddressRequired();
    error TokenSweepCallFailed(uint256 index, address receiver);

    event Deployed(address indexed receiver, bytes32 indexed salt, bytes32 initCodeHash);

    function sweep(bytes32[] calldata salts, bytes[] calldata initCodes) external override {
        uint256 length = salts.length;
        if (length == 0) {
            revert RecoveryInputsRequired();
        }
        if (length != initCodes.length) {
            revert RecoveryInputsLengthMismatch();
        }

        for (uint256 i = 0; i < length; ++i) {
            bytes calldata initCode = initCodes[i];
            if (initCode.length == 0) {
                revert InitCodeRequired(i);
            }

            bytes32 initCodeHash = keccak256(initCode);
            address receiver = _computeAddress(salts[i], initCodeHash);
            if (receiver.code.length == 0) {
                _deploy(receiver, salts[i], initCode, initCodeHash);
            }

            _sweepReceiver(i, receiver);
        }
    }

    function sweepERC20(
        bytes32[] calldata salts,
        bytes[] calldata initCodes,
        address token
    ) external override {
        if (token == address(0)) {
            revert TokenAddressRequired();
        }

        uint256 length = salts.length;
        if (length == 0) {
            revert RecoveryInputsRequired();
        }
        if (length != initCodes.length) {
            revert RecoveryInputsLengthMismatch();
        }

        for (uint256 i = 0; i < length; ++i) {
            bytes calldata initCode = initCodes[i];
            if (initCode.length == 0) {
                revert InitCodeRequired(i);
            }

            bytes32 initCodeHash = keccak256(initCode);
            address receiver = _computeAddress(salts[i], initCodeHash);
            if (receiver.code.length == 0) {
                _deploy(receiver, salts[i], initCode, initCodeHash);
            }

            _sweepReceiverERC20(i, receiver, token);
        }
    }

    function _deploy(
        address expectedReceiver,
        bytes32 salt,
        bytes calldata initCode,
        bytes32 initCodeHash
    ) private {
        address receiver;
        bytes memory creationCode = initCode;
        assembly {
            receiver := create2(0, add(creationCode, 0x20), mload(creationCode), salt)
        }

        if (receiver == address(0) || receiver != expectedReceiver) {
            revert Create2DeployFailed();
        }

        emit Deployed(receiver, salt, initCodeHash);
    }

    function _sweepReceiver(uint256 index, address receiver) private {
        (bool ok, ) = receiver.call(abi.encodeCall(IFixedCollectorReceiver.sweep, ()));
        if (!ok) {
            revert SweepCallFailed(index, receiver);
        }
    }

    function _sweepReceiverERC20(uint256 index, address receiver, address token) private {
        (bool ok, ) = receiver.call(abi.encodeCall(IFixedCollectorERC20Receiver.sweepERC20, (token)));
        if (!ok) {
            revert TokenSweepCallFailed(index, receiver);
        }
    }

    function _computeAddress(bytes32 salt, bytes32 initCodeHash) private view returns (address) {
        bytes32 digest = keccak256(
            abi.encodePacked(bytes1(0xff), address(this), salt, initCodeHash)
        );
        return address(uint160(uint256(digest)));
    }
}
