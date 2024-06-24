import React, { useEffect } from "react";
import { GetRoleRSOP } from "../services/roleService";

interface Props {
  principalId: string;
}

const RSOPPanel = ({ principalId }: Props) => {
  useEffect(() => {
    console.log("RSOPPanel");
    const { request, cancel } = GetRoleRSOP(principalId);

    request.then((response) => {
      console.log(response.data);
    });

    return () => {
      cancel();
    };
  }, []);

  return <div>RSOPPanel</div>;
};

export default RSOPPanel;
