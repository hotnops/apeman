import { useEffect } from "react";
import { Button, useToast } from "@chakra-ui/react";
import { MdContentCopy } from "react-icons/md";

interface Props {
  object: any;
}

const Export = ({ object }: Props) => {
  const toast = useToast();

  useEffect(() => {
    console.log("Export component received object:", object);
  }, [object]);

  const handleClick = () => {
    console.log("Object in handleClick:", object);
    navigator.clipboard.writeText(JSON.stringify(object, null, 2));

    toast({
      title: "Copied to clipboard",
      status: "success",
      duration: 2000,
      isClosable: true,
    });
  };

  return (
    <Button onClick={handleClick}>
      <MdContentCopy />
    </Button>
  );
};

export default Export;
