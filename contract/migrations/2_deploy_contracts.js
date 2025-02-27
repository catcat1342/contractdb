const ContractDB = artifacts.require("ContractDB");
const ContractDBMulti = artifacts.require("ContractDBMulti")

module.exports = function (deployer) {
    deployer.deploy(ContractDB);
    deployer.deploy(ContractDBMulti);
};
