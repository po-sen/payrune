// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

contract Create2ReceiverFactory {
    error Create2DeployFailed();
    error PostDeployCallFailed();

    event Deployed(address indexed receiver, bytes32 indexed salt, bytes32 initCodeHash);

    function deploy(bytes32 salt, bytes calldata initCode) public returns (address receiver) {
        bytes memory creationCode = initCode;
        assembly {
            receiver := create2(0, add(creationCode, 0x20), mload(creationCode), salt)
        }

        if (receiver == address(0)) {
            revert Create2DeployFailed();
        }

        emit Deployed(receiver, salt, keccak256(initCode));
    }

    function deployAndCall(
        bytes32 salt,
        bytes calldata initCode,
        bytes calldata callData
    ) external returns (address receiver, bytes memory returnData) {
        receiver = deploy(salt, initCode);
        (bool ok, bytes memory data) = receiver.call(callData);
        if (!ok) {
            revert PostDeployCallFailed();
        }
        return (receiver, data);
    }

    function computeAddress(bytes32 salt, bytes32 initCodeHash) external view returns (address) {
        return _computeAddress(address(this), salt, initCodeHash);
    }

    function _computeAddress(
        address deployer,
        bytes32 salt,
        bytes32 initCodeHash
    ) private pure returns (address) {
        bytes32 digest = keccak256(
            abi.encodePacked(bytes1(0xff), deployer, salt, initCodeHash)
        );
        return address(uint160(uint256(digest)));
    }
}
