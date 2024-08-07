/*
Copyright 2024 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import { CompareResult, VitessRefs } from "@/types";
import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";

import MacroBenchmarkTable from "@/common/MacroBenchmarkTable";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import useApiCall from "@/utils/Hook";
import { errorApi, formatCompareResult, formatGitRef } from "@/utils/Utils";
import { PlusCircledIcon } from "@radix-ui/react-icons";
import ForeignKeysHero from "./components/ForeignKeysHero";

export const formatTitle = (gitRef: string, vitessRefs: VitessRefs): string => {
  let title = formatGitRef(gitRef);
  vitessRefs.branches.forEach((branch) => {
    if (branch.commit_hash.match(gitRef)) {
      title = branch.name;
    }
  });
  vitessRefs.tags.forEach((branch) => {
    if (branch.commit_hash.match(gitRef)) {
      title = branch.name;
    }
  });
  return title;
};

export default function ForeignKeys() {
  const urlParams = new URLSearchParams(window.location.search);

  const [gitRef, setGitRef] = useState<string>(urlParams.get("sha") || "");
  const [workload, setWorkload] = useState<{ old: string; new: string }>({
    old: urlParams.get("oldWorkload") || "",
    new: urlParams.get("newWorkload") || "",
  });

  const { data: vitessRefs } = useApiCall<VitessRefs>(
    `${import.meta.env.VITE_API_URL}vitess/refs`
  );

  const {
    data: data,
    isLoading: isMacrobenchLoading,
    error: macrobenchError,
  } = useApiCall<CompareResult>(
    `${import.meta.env.VITE_API_URL}fk/compare?sha=${gitRef}&newWorkload=${
      workload.new
    }&oldWorkload=${workload.old}`
  );

  let formattedData = data !== null ? formatCompareResult(data) : null;

  const navigate = useNavigate();

  useEffect(() => {
    navigate(
      `?sha=${gitRef}&oldWorkload=${workload.old}&newWorkload=${workload.new}`
    );
  }, [gitRef, workload.old, workload.new]);

  return (
    <>
      {vitessRefs && (
        <ForeignKeysHero
          gitRef={gitRef}
          setGitRef={setGitRef}
          workload={workload}
          setWorkload={setWorkload}
          vitessRefs={vitessRefs}
        />
      )}

      {macrobenchError && (
        <div className="text-red-500 text-center my-2">{macrobenchError}</div>
      )}

      {(formattedData === null || data === null) && !isMacrobenchLoading && (
        <div className="text-red-500 text-center my-2">{errorApi}</div>
      )}

      <section className="flex flex-col items-center">
        {isMacrobenchLoading && (
          <>
            {[...Array(8)].map((_, index) => {
              return (
                <div key={index} className="w-[80vw] xl:w-[60vw] my-12">
                  <Skeleton className="h-[852px]"></Skeleton>
                </div>
              );
            })}
          </>
        )}
        {!isMacrobenchLoading &&
          formattedData &&
          data !== null &&
          vitessRefs && (
            <>
              <div className="w-[80vw] xl:w-[60vw] my-12">
                <Card className="border-border">
                  <CardHeader className="flex flex-col gap-4 md:gap-0 md:flex-row justify-between pt-6">
                    <CardTitle className="text-2xl md:text-4xl text-primary">
                      {gitRef == "" ? "N/A" : <Link
                          to={`https://github.com/vitessio/vitess/commit/${gitRef}`}
                          target="__blank"
                      >
                        {formatTitle(gitRef, vitessRefs)}
                      </Link>}
                    </CardTitle>
                    <Button
                      variant="outline"
                      size="sm"
                      className="h-8 w-fit border-dashed mt-4 md:mt-0"
                      disabled
                    >
                      <PlusCircledIcon className="mr-2 h-4 w-4 text-primary" />
                      {/* <Link
                          to={`/macrobench/queries/compare?ltag=${gitRef.old}&rtag=${gitRef.new}&workload=${macro.workload}`}
                        > */}
                      See Query Plan {/* </Link> */}
                    </Button>
                  </CardHeader>
                  <CardContent className="w-full p-0">
                    <MacroBenchmarkTable
                      data={formattedData}
                      new={workload.new}
                      old={workload.old}
                      isGitRef={false}
                    />
                  </CardContent>
                </Card>
              </div>
            </>
          )}
      </section>
    </>
  );
}
