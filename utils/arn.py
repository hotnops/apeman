import fnmatch


class Arn:

    def __init__(self):
        self.partition = "aws"
        self.service = None
        self.region = None
        self.account_id = None
        self.resource = None
        self.is_wildcard = False

    @classmethod
    def fromstring(cls, arn_string: str):
        arn = cls()
        if arn_string == "*":
            arn.partition = "*"
            arn.service = "*"
            arn.region = "*"
            arn.account_id = "*"
            arn.resource = "*"
            arn.is_wildcard = True

        elif Arn.is_arn(arn_string):
            arn.parse_arn(arn_string)

        return arn

    @staticmethod
    def is_arn(arn_arg: str):
        if arn_arg.startswith("arn") and arn_arg.count(":") >= 5:
            return True
        return False

    def get_account_id(self):
        return self.account_id

    def like(self, dest_arn_str: str) -> bool:
        if not Arn.is_arn(dest_arn_str):
            return False

        dest_arn = Arn.fromstring(dest_arn_str)

        if self.partition != dest_arn.partition:
            return False
        if not fnmatch.fnmatch(self.partition, dest_arn.partition):
            return False
        if not fnmatch.fnmatch(self.service, dest_arn.service):
            return False
        if not fnmatch.fnmatch(self.region, dest_arn.region):
            return False
        if not fnmatch.fnmatch(self.account_id, dest_arn.account_id):
            return False
        if not fnmatch.fnmatch(self.resource, dest_arn.resource):
            return False
        return True

    def parse_arn(self, arn: str):
        arn, self.partition, self.service, self.region, self.account_id, self.resource = arn.split(":", 5)

    def __str__(self):
        if self.is_wildcard:
            return "*"
        else:
            arn_string = f"arn:{self.partition}:{self.service}:{self.region}:{self.account_id}:{self.resource}"
            return arn_string

    def to_regex(self):
        if self.resource == "root":
            return str(self).replace("root", ".*")
        return str(self).replace("*", ".*")
